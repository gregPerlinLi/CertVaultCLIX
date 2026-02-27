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

// certDetailMode tracks whether we are showing cert details or inline analysis.
type certDetailMode int

const (
	certDetailNormal   certDetailMode = iota
	certDetailAnalysis                // showing analysis result
)

// CertDetail shows detailed information about an SSL certificate.
type CertDetail struct {
	Cert        *api.SSLCert
	client      *api.Client
	mode        certDetailMode
	spinner     components.Spinner
	resultVP    viewport.Model
	hasResult   bool
	analysisErr string
	width       int
	height      int
}

// NewCertDetail creates a new SSL cert detail view.
func NewCertDetail(cert *api.SSLCert, client *api.Client) CertDetail {
	vp := viewport.New(80, 20)
	return CertDetail{
		Cert:     cert,
		client:   client,
		spinner:  components.NewSpinner(),
		resultVP: vp,
	}
}

// SetSize updates dimensions.
func (c *CertDetail) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.resultVP.Width = width
	vpH := height - 5
	if vpH < 4 {
		vpH = 4
	}
	c.resultVP.Height = vpH
}

// Init does nothing.
func (c *CertDetail) Init() tea.Cmd { return nil }

// IsAnalysisMode returns true when the inline analysis result is being shown.
func (c *CertDetail) IsAnalysisMode() bool { return c.mode == certDetailAnalysis }

// Update handles messages.
func (c *CertDetail) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if c.spinner.IsActive() {
			return c.spinner.Update(msg)
		}
		switch c.mode {
		case certDetailAnalysis:
			switch msg.String() {
			case "esc":
				c.mode = certDetailNormal
				c.hasResult = false
				c.analysisErr = ""
				return nil
			case "up", "k", "down", "j", "pgup", "pgdown":
				var vpCmd tea.Cmd
				c.resultVP, vpCmd = c.resultVP.Update(msg)
				return vpCmd
			}
		case certDetailNormal:
			if msg.String() == "a" {
				return c.startAnalysis()
			}
		}
	case tea.MouseMsg:
		if c.mode == certDetailAnalysis {
			var vpCmd tea.Cmd
			c.resultVP, vpCmd = c.resultVP.Update(msg)
			return vpCmd
		}
	case inlineAnalysisMsg:
		c.spinner.Stop()
		c.mode = certDetailAnalysis
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

func (c *CertDetail) startAnalysis() tea.Cmd {
	uuid := c.Cert.UUID
	client := c.client
	vpWidth := c.resultVP.Width
	spinCmd := c.spinner.Start("Analyzing...")
	return tea.Batch(spinCmd, func() tea.Msg {
		ctx := context.Background()
		certContent, err := client.GetUserSSLCert(ctx, uuid)
		if err != nil {
			return inlineAnalysisMsg{err: err.Error()}
		}
		analysis, err := client.AnalyzeCert(ctx, certContent.Certificate)
		if err != nil {
			return inlineAnalysisMsg{err: err.Error()}
		}
		return inlineAnalysisMsg{result: formatCertAnalysis(analysis, vpWidth)}
	})
}

// View renders the cert detail.
func (c *CertDetail) View() string {
	if c.Cert == nil {
		return tui.MutedStyle.Render("No certificate selected.")
	}

	if c.mode == certDetailAnalysis {
		return c.viewAnalysis()
	}

	cert := c.Cert
	var sb strings.Builder

	sb.WriteString(tui.TitleStyle.Render("📜 SSL Certificate Details"))
	sb.WriteString("\n\n")

	daysLeft := parseDaysLeft(cert.NotAfter)
	expStyle := tui.ExpiryStyle(daysLeft)

	rows := []struct{ label, value string }{
		{"UUID", cert.UUID},
		{"CA UUID", cert.CaUUID},
		{"Owner", cert.Owner},
		{"Valid From", cert.NotBefore},
		{"Valid Until", expStyle.Render(cert.NotAfter) + fmt.Sprintf("  (%d days)", daysLeft)},
		{"Comment", cert.Comment},
		{"Created At", cert.CreatedAt},
		{"Modified At", cert.ModifiedAt},
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

func (c *CertDetail) viewAnalysis() string {
	var sb strings.Builder
	sb.WriteString(tui.TitleStyle.Render("📜 SSL Certificate Analysis"))
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
