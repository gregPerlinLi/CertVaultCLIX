package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
	tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
)

// CADetail shows detailed information about a CA certificate.
type CADetail struct {
CA     *api.CACert
width  int
height int
}

// NewCADetail creates a new CA detail view.
func NewCADetail(ca *api.CACert) CADetail {
return CADetail{CA: ca}
}

// SetSize updates dimensions.
func (c *CADetail) SetSize(width, height int) {
c.width = width
c.height = height
}

// Init does nothing for this static view.
func (c *CADetail) Init() tea.Cmd { return nil }

// Update handles navigation.
func (c *CADetail) Update(msg tea.Msg) tea.Cmd { return nil }

// View renders the CA detail.
func (c *CADetail) View() string {
if c.CA == nil {
return tui.MutedStyle.Render("No CA selected.")
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
{"Type", ca.CAType()},
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
sb.WriteString(tui.HelpStyle.Render("esc: back"))
return sb.String()
}
