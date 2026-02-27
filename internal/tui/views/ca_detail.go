package views

import (
	"fmt"
	"strings"
	"time"

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

	daysLeft := int(time.Until(ca.NotAfter).Hours() / 24)
	expStyle := tui.ExpiryStyle(daysLeft)

	rows := []struct{ label, value string }{
		{"UUID", ca.UUID},
		{"Common Name (CN)", ca.CN},
		{"Country", ca.Country},
		{"Province", ca.Province},
		{"City", ca.City},
		{"Organization", ca.Org},
		{"Org Unit", ca.OrgUnit},
		{"Email", ca.Email},
		{"Algorithm", ca.Algorithm},
		{"Key Size", fmt.Sprintf("%d bits", ca.KeySize)},
		{"Valid From", ca.NotBefore.Format("2006-01-02 15:04:05")},
		{"Valid Until", expStyle.Render(ca.NotAfter.Format("2006-01-02 15:04:05")) + "  " + expStyle.Render(fmt.Sprintf("(%d days)", daysLeft))},
		{"Is CA", fmt.Sprintf("%v", ca.IsCA)},
		{"Parent CA", ca.ParentCA},
		{"Available", fmt.Sprintf("%v", ca.Available)},
		{"Comment", ca.Comment},
		{"Owner", ca.Owner},
		{"Created At", ca.CreatedAt.Format("2006-01-02 15:04:05")},
		{"Updated At", ca.UpdatedAt.Format("2006-01-02 15:04:05")},
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
