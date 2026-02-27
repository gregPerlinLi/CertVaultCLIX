package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
	tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
)

// CertDetail shows detailed information about an SSL certificate.
type CertDetail struct {
	Cert   *api.SSLCert
	width  int
	height int
}

// NewCertDetail creates a new SSL cert detail view.
func NewCertDetail(cert *api.SSLCert) CertDetail {
	return CertDetail{Cert: cert}
}

// SetSize updates dimensions.
func (c *CertDetail) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// Init does nothing.
func (c *CertDetail) Init() tea.Cmd { return nil }

// Update handles navigation.
func (c *CertDetail) Update(msg tea.Msg) tea.Cmd { return nil }

// View renders the cert detail.
func (c *CertDetail) View() string {
	if c.Cert == nil {
		return tui.MutedStyle.Render("No certificate selected.")
	}
	cert := c.Cert
	var sb strings.Builder

	sb.WriteString(tui.TitleStyle.Render("📜 SSL Certificate Details"))
	sb.WriteString("\n\n")

	daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)
	expStyle := tui.ExpiryStyle(daysLeft)

	sans := strings.Join(cert.SANs, ", ")
	if sans == "" {
		sans = "(none)"
	}

	rows := []struct{ label, value string }{
		{"UUID", cert.UUID},
		{"Common Name (CN)", cert.CN},
		{"Country", cert.Country},
		{"Province", cert.Province},
		{"City", cert.City},
		{"Organization", cert.Org},
		{"Org Unit", cert.OrgUnit},
		{"Email", cert.Email},
		{"SANs", sans},
		{"Algorithm", cert.Algorithm},
		{"Key Size", fmt.Sprintf("%d bits", cert.KeySize)},
		{"Valid From", cert.NotBefore.Format("2006-01-02 15:04:05")},
		{"Valid Until", expStyle.Render(cert.NotAfter.Format("2006-01-02 15:04:05")) + "  " + expStyle.Render(fmt.Sprintf("(%d days)", daysLeft))},
		{"CA", cert.CA},
		{"Comment", cert.Comment},
		{"Owner", cert.Owner},
		{"Created At", cert.CreatedAt.Format("2006-01-02 15:04:05")},
		{"Updated At", cert.UpdatedAt.Format("2006-01-02 15:04:05")},
	}

	for _, row := range rows {
		if row.value == "" {
			continue
		}
		label := tui.KeyStyle.Render(fmt.Sprintf("%-22s", row.label+":"))
		sb.WriteString(label + " " + row.value + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(tui.HelpStyle.Render("esc: back"))
	return sb.String()
}
