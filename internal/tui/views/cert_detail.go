package views

import (
	"fmt"
	"strings"

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
sb.WriteString(tui.HelpStyle.Render("esc: back"))
return sb.String()
}
