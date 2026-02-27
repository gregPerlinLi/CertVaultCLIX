package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
	tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/components"
)

// ToolsMode indicates which tool is active.
type ToolsMode int

const (
	ToolsModeMenu ToolsMode = iota
	ToolsModeAnalyzeCert
	ToolsModeAnalyzeKey
	ToolsModeConvert
)

// Tools is the certificate tools view.
type Tools struct {
	client   *api.Client
	mode     ToolsMode
	menuIdx  int
	input    textarea.Model
	result   string
	spinner  components.Spinner
	err      string
	width    int
	height   int
}

var toolsMenuItems = []string{
	"Analyze Certificate",
	"Analyze Private Key",
	"Convert PEM → DER",
	"Convert DER → PEM",
}

// NewTools creates a new tools view.
func NewTools(client *api.Client) Tools {
	ta := textarea.New()
	ta.Placeholder = "Paste PEM/DER content here..."
	ta.SetWidth(60)
	ta.SetHeight(10)
	return Tools{
		client:  client,
		input:   ta,
		spinner: components.NewSpinner(),
	}
}

// SetSize updates dimensions.
func (t *Tools) SetSize(width, height int) {
	t.width = width
	t.height = height
}

// Init initializes.
func (t *Tools) Init() tea.Cmd { return nil }

// Update handles messages.
func (t *Tools) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if t.spinner.IsActive() {
			return t.spinner.Update(msg)
		}
		switch t.mode {
		case ToolsModeMenu:
			switch msg.String() {
			case "up", "k":
				if t.menuIdx > 0 {
					t.menuIdx--
				}
			case "down", "j":
				if t.menuIdx < len(toolsMenuItems)-1 {
					t.menuIdx++
				}
			case "enter":
				switch t.menuIdx {
				case 0:
					t.mode = ToolsModeAnalyzeCert
					t.input.Focus()
					t.result = ""
					t.err = ""
				case 1:
					t.mode = ToolsModeAnalyzeKey
					t.input.Focus()
					t.result = ""
					t.err = ""
				case 2, 3:
					t.mode = ToolsModeConvert
					t.input.Focus()
					t.result = ""
					t.err = ""
				}
			}
		default:
			switch msg.String() {
			case "esc":
				t.mode = ToolsModeMenu
				t.input.Blur()
				t.result = ""
				t.err = ""
			case "ctrl+s", "ctrl+enter":
				return t.runTool()
			case "ctrl+l":
				// Clear the textarea content
				t.input.Reset()
				return nil
			default:
				var cmd tea.Cmd
				t.input, cmd = t.input.Update(msg)
				return cmd
			}
		}
	case toolResultMsg:
		t.spinner.Stop()
		t.result = msg.result
		t.err = msg.err
		return nil
	}
	return t.spinner.Update(msg)
}

type toolResultMsg struct {
	result string
	err    string
}

func (t *Tools) runTool() tea.Cmd {
	content := t.input.Value()
	mode := t.mode
	menuIdx := t.menuIdx
	cmd := t.spinner.Start("Processing...")
	return tea.Batch(cmd, func() tea.Msg {
		ctx := context.Background()
		switch mode {
		case ToolsModeAnalyzeCert:
			analysis, err := t.client.AnalyzeCert(ctx, content)
			if err != nil {
				return toolResultMsg{err: err.Error()}
			}
			return toolResultMsg{result: formatCertAnalysis(analysis)}
		case ToolsModeAnalyzeKey:
			analysis, err := t.client.AnalyzePrivKey(ctx, content, "")
			if err != nil {
				return toolResultMsg{err: err.Error()}
			}
			return toolResultMsg{result: formatPrivKeyAnalysis(analysis)}
		case ToolsModeConvert:
			switch menuIdx {
			case 2: // PEM to DER
				result, err := t.client.ConvertPEMtoDER(ctx, content)
				if err != nil {
					return toolResultMsg{err: err.Error()}
				}
				return toolResultMsg{result: result.Data}
			case 3: // DER to PEM
				result, err := t.client.ConvertDERtoPEM(ctx, content)
				if err != nil {
					return toolResultMsg{err: err.Error()}
				}
				return toolResultMsg{result: result.Data}
			}
		}
		return toolResultMsg{err: "unknown tool"}
	})
}

func formatCertAnalysis(a *api.CertAnalysis) string {
	var sb strings.Builder
	sb.WriteString("=== Certificate Analysis ===\n\n")
	sb.WriteString("Algorithm: " + a.Algorithm + "\n")
	sb.WriteString("Is CA: " + boolStr(a.IsCA) + "\n")
	sb.WriteString("Not Before: " + a.NotBefore + "\n")
	sb.WriteString("Not After: " + a.NotAfter + "\n")
	sb.WriteString("Fingerprint: " + a.Fingerprint + "\n")
	if len(a.SANs) > 0 {
		sb.WriteString("SANs: " + strings.Join(a.SANs, ", ") + "\n")
	}
	return sb.String()
}

func formatPrivKeyAnalysis(a *api.PrivKeyAnalysis) string {
	return "Algorithm: " + a.Algorithm + "\nKey Size: " + itoa(a.KeySize) + " bits"
}

func boolStr(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

// View renders the tools view.
func (t *Tools) View() string {
	var sb strings.Builder
	sb.WriteString(tui.TitleStyle.Render("🛠 Tools"))
	sb.WriteString("\n\n")

	if t.mode == ToolsModeMenu {
		for i, item := range toolsMenuItems {
			if i == t.menuIdx {
				sb.WriteString(tui.SelectedStyle.Render("▶ " + item))
			} else {
				sb.WriteString(tui.NormalStyle.Render("  " + item))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
		sb.WriteString(tui.HelpStyle.Render("↑/↓: select • enter: open tool"))
		return sb.String()
	}

	sb.WriteString(tui.SubtitleStyle.Render(toolsMenuItems[t.menuIdx]))
	sb.WriteString("\n\n")
	sb.WriteString(tui.NormalStyle.Render("Input:"))
	sb.WriteString("\n")
	sb.WriteString(t.input.View())
	sb.WriteString("\n\n")

	if t.spinner.IsActive() {
		sb.WriteString(t.spinner.View())
	} else if t.result != "" {
		sb.WriteString(tui.SuccessStyle.Render("Result:"))
		sb.WriteString("\n")
		sb.WriteString(tui.BorderStyle.Render(t.result))
	} else if t.err != "" {
		sb.WriteString(tui.DangerStyle.Render("Error: " + t.err))
	}

	sb.WriteString("\n\n")
	sb.WriteString(tui.HelpStyle.Render("ctrl+s: run • ctrl+l: clear • esc: back"))
	return sb.String()
}

// IsAtRoot returns true when the Tools view is showing the top-level menu.
func (t *Tools) IsAtRoot() bool {
return t.mode == ToolsModeMenu
}
