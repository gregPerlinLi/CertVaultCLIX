package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
	tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/components"
)

// CARequestedMsg is sent when a new CA cert has been issued.
type CARequestedMsg struct {
	CA  *api.CACert
	Err error
}

// CARequest is the form for requesting a new CA certificate.
// It uses a viewport so the form is always scrollable — even in small terminals.
type CARequest struct {
	client   *api.Client
	fields   []*components.FormField
	form     components.Form
	viewport viewport.Model
	spinner  components.Spinner
	err      string
	width    int
	height   int
}

// NewCARequest creates a new CA request form.
func NewCARequest(client *api.Client) CARequest {
	fields := []*components.FormField{
		{Label: "Parent CA UUID", Placeholder: "Leave empty for Root CA"},
		{Label: "Allow Sub-CA", Placeholder: "true or false (enable intermediate CA)"},
		{Label: "Common Name (CN)", Placeholder: "e.g. My Root CA"},
		{Label: "Country", Placeholder: "e.g. CN"},
		{Label: "Province", Placeholder: "e.g. Guangdong"},
		{Label: "City", Placeholder: "e.g. Canton"},
		{Label: "Organization", Placeholder: "e.g. Acme Corp"},
		{Label: "Org Unit", Placeholder: "e.g. IT Department"},
		{Label: "Algorithm", Placeholder: "RSA, EC, or ED25519"},
		{Label: "Key Size", Placeholder: "2048/4096 (RSA) • 256/384 (EC) • leave empty for ED25519"},
		{Label: "Expire Days", Placeholder: "e.g. 3650"},
		{Label: "Comment", Placeholder: "Optional comment"},
	}
	f := components.NewForm("🔒 Request CA Certificate", fields)
	vp := viewport.New(80, 20)
	vp.SetContent(f.View())
	return CARequest{
		client:   client,
		fields:   fields,
		form:     f,
		viewport: vp,
		spinner:  components.NewSpinner(),
	}
}

// SetSize updates dimensions.
func (c *CARequest) SetSize(width, height int) {
	c.width = width
	c.height = height
	vpHeight := height - 3
	if vpHeight < 3 {
		vpHeight = 3
	}
	c.viewport.Width = width
	c.viewport.Height = vpHeight
	c.refreshViewport()
}

// Init initializes the form.
func (c *CARequest) Init() tea.Cmd {
	c.form.Reset()
	c.err = ""
	c.viewport.GotoTop()
	c.refreshViewport()
	return textinput.Blink
}

// refreshViewport scrolls to the focused field.
func (c *CARequest) refreshViewport() {
	c.viewport.SetContent(c.form.View())
	focused := c.form.FocusedIndex()
	targetLine := formTitleLines + focused*linesPerField
	offset := targetLine - c.viewport.Height/2
	if offset < 0 {
		offset = 0
	}
	c.viewport.SetYOffset(offset)
}

// Update handles messages.
func (c *CARequest) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if c.spinner.IsActive() {
			return c.spinner.Update(msg)
		}
		switch msg.String() {
		case "enter":
			if c.form.FocusedIndex() == len(c.fields)-1 {
				return c.submit()
			}
			formCmd := c.form.Update(msg)
			c.refreshViewport()
			return formCmd
		default:
			formCmd := c.form.Update(msg)
			c.refreshViewport()
			return formCmd
		}

	case tea.MouseMsg:
		var vpCmd tea.Cmd
		c.viewport, vpCmd = c.viewport.Update(msg)
		return vpCmd

	case CARequestedMsg:
		c.spinner.Stop()
		if msg.Err != nil {
			if isUnauthorized(msg.Err) {
				return func() tea.Msg { return SessionExpiredMsg{} }
			}
			c.err = msg.Err.Error()
		}
		return nil
	}

	return c.spinner.Update(msg)
}

func (c *CARequest) submit() tea.Cmd {
	c.err = ""
	parentCaUUID := c.form.Value(0)
	allowSubCaStr := strings.ToLower(strings.TrimSpace(c.form.Value(1)))
	cn := c.form.Value(2)
	country := c.form.Value(3)
	province := c.form.Value(4)
	city := c.form.Value(5)
	org := c.form.Value(6)
	ou := c.form.Value(7)
	algo := c.form.Value(8)
	keySizeStr := c.form.Value(9)
	expireDaysStr := c.form.Value(10)
	comment := c.form.Value(11)

	if cn == "" {
		c.err = "Common Name is required"
		return nil
	}

	allowSubCa := allowSubCaStr == "true" || allowSubCaStr == "yes" || allowSubCaStr == "1"

	keySize := 0
	if keySizeStr != "" {
		fmt.Sscanf(keySizeStr, "%d", &keySize)
	}

	expireDays := 3650
	fmt.Sscanf(expireDaysStr, "%d", &expireDays)

	if algo == "" {
		algo = "RSA"
	}
	if keySize == 0 && (algo == "RSA" || algo == "EC") {
		if algo == "RSA" {
			keySize = 2048
		} else {
			keySize = 256
		}
	}

	req := api.RequestCACertRequest{
		CaUUID:             parentCaUUID,
		AllowSubCa:         allowSubCa,
		Algorithm:          algo,
		KeySize:            keySize,
		CommonName:         cn,
		Country:            country,
		Province:           province,
		City:               city,
		Organization:       org,
		OrganizationalUnit: ou,
		Expiry:             expireDays,
		Comment:            comment,
	}

	cmd := c.spinner.Start("Requesting CA certificate...")
	return tea.Batch(cmd, func() tea.Msg {
		ca, err := c.client.RequestAdminCA(context.Background(), req)
		return CARequestedMsg{CA: ca, Err: err}
	})
}

// View renders the CA request form inside a viewport.
func (c *CARequest) View() string {
	var sb strings.Builder

	sb.WriteString(c.viewport.View())
	sb.WriteString("\n")

	if c.spinner.IsActive() {
		sb.WriteString(c.spinner.View())
		sb.WriteString("\n")
	}
	if c.err != "" {
		sb.WriteString(tui.DangerStyle.Render("✗ " + c.err))
		sb.WriteString("\n")
	}
	pct := c.viewport.ScrollPercent()
	scrollInfo := ""
	if pct >= 0 && pct <= 1 {
		scrollInfo = fmt.Sprintf(" [%.0f%%]", pct*100)
	}
	sb.WriteString(tui.HelpStyle.Render(
		"tab/↓: next field • shift+tab/↑: prev • enter (last): submit • scroll: mouse wheel" + scrollInfo,
	))
	return sb.String()
}
