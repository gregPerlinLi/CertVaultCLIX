package views

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	client       *api.Client
	mode         ToolsMode
	menuIdx      int
	input        textarea.Model
	resultVP     viewport.Model
	contentWidth int // viewport usable width, set by SetSize
	hasResult    bool
	resultFocus  bool // true = result viewport has keyboard focus
	spinner      components.Spinner
	err          string
	width        int
	height       int
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
	ta.Placeholder = "Paste PEM content here (e.g. -----BEGIN CERTIFICATE-----)..."
	ta.SetWidth(60)
	ta.SetHeight(toolsInputHeight)
	vp := viewport.New(80, 20)
	return Tools{
		client:   client,
		input:    ta,
		resultVP: vp,
		spinner:  components.NewSpinner(),
	}
}

// toolsInputHeight is the number of lines the input textarea occupies.
const toolsInputHeight = 6

// SetSize updates dimensions.
func (t *Tools) SetSize(width, height int) {
	t.width = width
	t.height = height
	t.contentWidth = width - 2
	if t.contentWidth < 20 {
		t.contentWidth = 20
	}
	t.input.SetWidth(width - 2)
	// Result viewport: remaining height after title (2) + subtitle (1) + input label (1) + textarea + help (2)
	vpHeight := height - (2 + 1 + 1 + toolsInputHeight + 1 + 2)
	if vpHeight < 4 {
		vpHeight = 4
	}
	t.resultVP.Width = width
	t.resultVP.Height = vpHeight
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
					t.hasResult = false
					t.resultFocus = false
					t.err = ""
				case 1:
					t.mode = ToolsModeAnalyzeKey
					t.input.Focus()
					t.hasResult = false
					t.resultFocus = false
					t.err = ""
				case 2, 3:
					t.mode = ToolsModeConvert
					t.input.Focus()
					t.hasResult = false
					t.resultFocus = false
					t.err = ""
				}
			}
		default:
			switch msg.String() {
			case "esc":
				t.mode = ToolsModeMenu
				t.input.Blur()
				t.hasResult = false
				t.resultFocus = false
				t.err = ""
			case "tab":
				// Toggle focus between input and result viewport.
				if t.hasResult {
					t.resultFocus = !t.resultFocus
					if t.resultFocus {
						t.input.Blur()
					} else {
						t.input.Focus()
					}
				}
			case "ctrl+s", "ctrl+enter":
				return t.runTool()
			case "ctrl+l":
				t.input.Reset()
				t.hasResult = false
				t.resultFocus = false
				t.err = ""
				t.input.Focus()
				return nil
			case "up", "k", "down", "j", "pgup", "pgdown":
				// Scroll result viewport when it has focus.
				if t.resultFocus {
					var vpCmd tea.Cmd
					t.resultVP, vpCmd = t.resultVP.Update(msg)
					return vpCmd
				}
				var cmd tea.Cmd
				t.input, cmd = t.input.Update(msg)
				return cmd
			default:
				if !t.resultFocus {
					var cmd tea.Cmd
					t.input, cmd = t.input.Update(msg)
					return cmd
				}
			}
		}
	case tea.MouseMsg:
		if t.hasResult {
			var vpCmd tea.Cmd
			t.resultVP, vpCmd = t.resultVP.Update(msg)
			return vpCmd
		}
	case toolResultMsg:
		t.spinner.Stop()
		t.err = msg.err
		if msg.result != "" {
			t.hasResult = true
			t.resultVP.SetContent(msg.result)
			t.resultVP.GotoTop()
			t.resultFocus = true
			t.input.Blur()
		}
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
	vpWidth := t.contentWidth
	cmd := t.spinner.Start("Processing...")
	return tea.Batch(cmd, func() tea.Msg {
		ctx := context.Background()
		switch mode {
		case ToolsModeAnalyzeCert:
			analysis, err := t.client.AnalyzeCert(ctx, base64.StdEncoding.EncodeToString([]byte(content)))
			if err != nil {
				return toolResultMsg{err: err.Error()}
			}
			return toolResultMsg{result: formatCertAnalysis(analysis, vpWidth)}
		case ToolsModeAnalyzeKey:
			analysis, err := t.client.AnalyzePrivKey(ctx, base64.StdEncoding.EncodeToString([]byte(content)), "")
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

func formatCertAnalysis(a *api.CertAnalysis, maxWidth int) string {
	sectionStyle := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	keyStyle := lipgloss.NewStyle().Foreground(tui.ColorTextMuted)

	const keyWidth = 19 // "Key:              " including trailing space
	valueWidth := maxWidth - keyWidth
	if valueWidth < 20 {
		valueWidth = 20
	}
	indent := strings.Repeat(" ", keyWidth)

	field := func(key, value string) string {
		k := keyStyle.Render(fmt.Sprintf("%-18s", key+":"))
		v := wrapText(value, valueWidth, indent)
		return k + " " + v + "\n"
	}
	// expiryField colors only the Not After date.
	expiryField := func(key, dateStr string) string {
		k := keyStyle.Render(fmt.Sprintf("%-18s", key+":"))
		daysLeft := parseDaysLeft(dateStr)
		v := tui.ExpiryStyle(daysLeft).Render(dateStr)
		return k + " " + v + "\n"
	}

	var sb strings.Builder
	sb.WriteString(sectionStyle.Render("Certificate Analysis"))
	sb.WriteString("\n\n")

	if a.Subject != "" {
		sb.WriteString(field("Subject", a.Subject))
	}
	if a.Issuer != "" {
		sb.WriteString(field("Issuer", a.Issuer))
	}
	sb.WriteString(field("Not Before", a.NotBefore))
	sb.WriteString(expiryField("Not After", a.NotAfter))
	if a.SerialNumber != "" {
		sb.WriteString(field("Serial Number", a.SerialNumber))
	}
	if a.Fingerprint != "" {
		sb.WriteString(field("Fingerprint", a.Fingerprint))
	}
	sb.WriteString(field("Is CA", boolStr(a.IsCA)))
	sb.WriteString("\n")

	// Public Key section
	sb.WriteString(sectionStyle.Render("Public Key"))
	sb.WriteString("\n\n")
	sb.WriteString(field("Algorithm", a.Algorithm))
	if len(a.PublicKey) > 0 {
		keys := make([]string, 0, len(a.PublicKey))
		for k := range a.PublicKey {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := a.PublicKey[k]
			var vStr string
			switch val := v.(type) {
			case string:
				vStr = val
			case float64:
				vStr = fmt.Sprintf("%.0f", val)
			case bool:
				vStr = boolStr(val)
			default:
				b, _ := json.Marshal(v)
				vStr = string(b)
			}
			sb.WriteString(field(k, vStr))
		}
	}
	sb.WriteString("\n")

	// Extensions
	if len(a.Extensions) > 0 {
		sb.WriteString(sectionStyle.Render("Extensions"))
		sb.WriteString("\n\n")
		extKeys := make([]string, 0, len(a.Extensions))
		for k := range a.Extensions {
			extKeys = append(extKeys, k)
		}
		sort.Strings(extKeys)
		for _, k := range extKeys {
			sb.WriteString(field(k, a.Extensions[k]))
		}
		sb.WriteString("\n")
	} else if len(a.SANs) > 0 {
		sb.WriteString(sectionStyle.Render("Extensions"))
		sb.WriteString("\n\n")
		sb.WriteString(field("2.5.29.17", "SAN: "+strings.Join(a.SANs, ", ")))
		sb.WriteString("\n")
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
	sb.WriteString("\n")

	// Input area
	inputLabel := "Input:"
	if t.resultFocus {
		inputLabel = tui.MutedStyle.Render("Input: [tab: edit]")
	}
	sb.WriteString(inputLabel)
	sb.WriteString("\n")
	sb.WriteString(t.input.View())
	sb.WriteString("\n")

	// Spinner / error / result viewport
	if t.spinner.IsActive() {
		sb.WriteString(t.spinner.View())
		sb.WriteString("\n")
	} else if t.hasResult {
		focusHint := ""
		if !t.resultFocus {
			focusHint = tui.MutedStyle.Render(" [tab: scroll]")
		}
		sb.WriteString(tui.SuccessStyle.Render("Result:") + focusHint + "\n")
		sb.WriteString(t.resultVP.View())
		if t.resultVP.TotalLineCount() > t.resultVP.Height {
			pct := int(t.resultVP.ScrollPercent() * 100)
			sb.WriteString(tui.MutedStyle.Render(fmt.Sprintf(" %d%%", pct)))
		}
		sb.WriteString("\n")
	} else if t.err != "" {
		sb.WriteString(tui.DangerStyle.Render("Error: " + t.err))
		sb.WriteString("\n")
	}

	helpStr := "ctrl+s: run • ctrl+l: clear • esc: back"
	if t.hasResult && t.resultFocus {
		helpStr = "↑/↓: scroll • tab: edit input • ctrl+l: clear • esc: back"
	}
	sb.WriteString(tui.HelpStyle.Render(helpStr))
	return sb.String()
}

// IsAtRoot returns true when the Tools view is showing the top-level menu.
func (t *Tools) IsAtRoot() bool {
return t.mode == ToolsModeMenu
}
