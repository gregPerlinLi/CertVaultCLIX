package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
	tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/components"
)

// caDetailMode tracks whether we are showing CA details or inline analysis.
type caDetailMode int

const (
	caDetailNormal   caDetailMode = iota
	caDetailAnalysis              // showing analysis result
)

// CADetail shows detailed information about a CA certificate.
type CADetail struct {
	CA          *api.CACert
	client      *api.Client
	isAdmin     bool // use admin API for fetching cert
	mode        caDetailMode
	spinner     components.Spinner
	resultVP    viewport.Model
	hasResult   bool
	analysisErr string
	width       int
	height      int
}

// NewCADetail creates a new CA detail view.
// Pass isAdmin=true when used inside the Admin view (uses admin API endpoints).
func NewCADetail(ca *api.CACert, client *api.Client, isAdmin bool) CADetail {
	vp := viewport.New(80, 20)
	return CADetail{
		CA:       ca,
		client:   client,
		isAdmin:  isAdmin,
		spinner:  components.NewSpinner(),
		resultVP: vp,
	}
}

// SetSize updates dimensions.
func (c *CADetail) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.resultVP.Width = width
	vpH := height - 5
	if vpH < 4 {
		vpH = 4
	}
	c.resultVP.Height = vpH
}

// Init does nothing for this static view.
func (c *CADetail) Init() tea.Cmd { return nil }

// IsAnalysisMode returns true when the inline analysis result is being shown.
func (c *CADetail) IsAnalysisMode() bool { return c.mode == caDetailAnalysis }

// Update handles messages.
func (c *CADetail) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if c.spinner.IsActive() {
			return c.spinner.Update(msg)
		}
		switch c.mode {
		case caDetailAnalysis:
			switch msg.String() {
			case "esc":
				c.mode = caDetailNormal
				c.hasResult = false
				c.analysisErr = ""
				return nil
			case "up", "k", "down", "j", "pgup", "pgdown":
				var vpCmd tea.Cmd
				c.resultVP, vpCmd = c.resultVP.Update(msg)
				return vpCmd
			}
		case caDetailNormal:
			if msg.String() == "a" {
				return c.startAnalysis()
			}
		}
	case tea.MouseMsg:
		if c.mode == caDetailAnalysis {
			var vpCmd tea.Cmd
			c.resultVP, vpCmd = c.resultVP.Update(msg)
			return vpCmd
		}
	case inlineAnalysisMsg:
		c.spinner.Stop()
		c.mode = caDetailAnalysis
		c.analysisErr = msg.err
		if msg.result != "" {
			c.hasResult = true
			c.resultVP.SetContent(msg.result)
			c.resultVP.GotoTop()
		}
		return nil
	}
	return c.spinner.Update(msg)
}

func (c *CADetail) startAnalysis() tea.Cmd {
	uuid := c.CA.UUID
	client := c.client
	isAdmin := c.isAdmin
	vpWidth := c.resultVP.Width
	spinCmd := c.spinner.Start("Analyzing...")
	return tea.Batch(spinCmd, func() tea.Msg {
		ctx := context.Background()
		var certPEM string
		if isAdmin {
			certContent, err := client.GetAdminCACert(ctx, uuid, false, false)
			if err != nil {
				return inlineAnalysisMsg{err: err.Error()}
			}
			certPEM = certContent.Certificate
		} else {
			certContent, err := client.GetUserCACert(ctx, uuid, false, false)
			if err != nil {
				return inlineAnalysisMsg{err: err.Error()}
			}
			certPEM = certContent.Certificate
		}
		analysis, err := client.AnalyzeCert(ctx, certPEM)
		if err != nil {
			return inlineAnalysisMsg{err: err.Error()}
		}
		return inlineAnalysisMsg{result: formatCertAnalysis(analysis, vpWidth)}
	})
}

// View renders the CA detail.
func (c *CADetail) View() string {
	if c.CA == nil {
		return tui.MutedStyle.Render("No CA selected.")
	}

	if c.mode == caDetailAnalysis {
		return c.viewAnalysis()
	}

	ca := c.CA
	var sb strings.Builder

	sb.WriteString(tui.TitleStyle.Render("🔐 CA Certificate Details"))
	sb.WriteString("\n\n")

	daysLeft := parseDaysLeft(ca.NotAfter)
	expStyle := tui.ExpiryStyle(daysLeft)

	available := "No"
	if ca.Available {
		available = "Yes"
	}

	rows := []struct{ label, value string }{
		{"UUID", ca.UUID},
		{"Owner", ca.Owner},
		{"Type", tui.CATypeStyle(ca.CAType()).Render(ca.CAType())},
		{"Parent CA UUID", ca.ParentCa},
		{"Valid From", ca.NotBefore},
		{"Valid Until", expStyle.Render(ca.NotAfter) + fmt.Sprintf("  (%d days)", daysLeft)},
		{"Available", available},
		{"Comment", ca.Comment},
	}

	for _, row := range rows {
		if row.value == "" {
			continue
		}
		label := tui.KeyStyle.Render(fmt.Sprintf("%-22s", row.label+":"))
		sb.WriteString(label + " " + row.value + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(tui.HelpStyle.Render("a: analyze cert • esc: back"))
	return sb.String()
}

func (c *CADetail) viewAnalysis() string {
	var sb strings.Builder
	sb.WriteString(tui.TitleStyle.Render("🔐 CA Certificate Analysis"))
	sb.WriteString("\n\n")

	if c.spinner.IsActive() {
		sb.WriteString(c.spinner.View())
		sb.WriteString("\n")
	} else if c.hasResult {
		sb.WriteString(c.resultVP.View())
		if c.resultVP.TotalLineCount() > c.resultVP.Height {
			pct := int(c.resultVP.ScrollPercent() * 100)
			sb.WriteString(tui.MutedStyle.Render(fmt.Sprintf(" %d%%", pct)))
		}
		sb.WriteString("\n")
	} else if c.analysisErr != "" {
		sb.WriteString(tui.DangerStyle.Render("Error: " + c.analysisErr))
		sb.WriteString("\n")
	}

	sb.WriteString(tui.HelpStyle.Render("↑/↓: scroll • esc: back to details"))
	return sb.String()
}
