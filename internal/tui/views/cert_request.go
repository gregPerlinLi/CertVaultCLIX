package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
	tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/components"
)

// CertRequestedMsg is sent when a new SSL cert has been issued.
type CertRequestedMsg struct {
	Cert *api.SSLCert
	Err  error
}

// CertRequest is the form for requesting a new SSL certificate.
type CertRequest struct {
	client  *api.Client
	fields  []*components.FormField
	form    components.Form
	spinner components.Spinner
	err     string
	width   int
	height  int
}

// NewCertRequest creates a new cert request form.
func NewCertRequest(client *api.Client) CertRequest {
	fields := []*components.FormField{
		{Label: "CA UUID", Placeholder: "UUID of the CA to sign this cert"},
		{Label: "Common Name (CN)", Placeholder: "e.g. example.com"},
		{Label: "Country", Placeholder: "e.g. US"},
		{Label: "Province", Placeholder: "e.g. California"},
		{Label: "City", Placeholder: "e.g. San Francisco"},
		{Label: "Organization", Placeholder: "e.g. Acme Corp"},
		{Label: "SANs", Placeholder: "Comma-separated: example.com,*.example.com"},
		{Label: "Algorithm", Placeholder: "RSA or ECDSA"},
		{Label: "Key Size", Placeholder: "2048 or 4096 (RSA) / 256 or 384 (ECDSA)"},
		{Label: "Expire Days", Placeholder: "e.g. 365"},
		{Label: "Comment", Placeholder: "Optional comment"},
	}
	f := components.NewForm("📜 Request SSL Certificate", fields)
	return CertRequest{
		client:  client,
		fields:  fields,
		form:    f,
		spinner: components.NewSpinner(),
	}
}

// SetSize updates dimensions.
func (c *CertRequest) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// Init initializes the form.
func (c *CertRequest) Init() tea.Cmd {
	c.form.Reset()
	return textinput.Blink
}

// Update handles messages.
func (c *CertRequest) Update(msg tea.Msg) tea.Cmd {
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
		}
	case CertRequestedMsg:
		c.spinner.Stop()
		if msg.Err != nil {
			c.err = msg.Err.Error()
		}
		return nil
	}

	spinCmd := c.spinner.Update(msg)
	formCmd := c.form.Update(msg)
	return tea.Batch(spinCmd, formCmd)
}

func (c *CertRequest) submit() tea.Cmd {
	c.err = ""
	caUUID := c.form.Value(0)
	cn := c.form.Value(1)
	country := c.form.Value(2)
	province := c.form.Value(3)
	city := c.form.Value(4)
	org := c.form.Value(5)
	sansStr := c.form.Value(6)
	algo := c.form.Value(7)
	keySizeStr := c.form.Value(8)
	expireDaysStr := c.form.Value(9)
	comment := c.form.Value(10)

	if cn == "" || caUUID == "" {
		c.err = "CN and CA UUID are required"
		return nil
	}

	// Build SANs list as SubjectAltName structs (DNS_NAME)
	var sans []api.SubjectAltName
	if sansStr != "" {
		for _, s := range strings.Split(sansStr, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				sans = append(sans, api.SubjectAltName{Type: "DNS_NAME", Value: s})
			}
		}
	}

	keySize := 2048
	fmt.Sscanf(keySizeStr, "%d", &keySize)

	expireDays := 365
	fmt.Sscanf(expireDaysStr, "%d", &expireDays)

	if algo == "" {
		algo = "RSA"
	}

	req := api.RequestSSLCertRequest{
		CaUUID:             caUUID,
		Algorithm:          algo,
		KeySize:            keySize,
		CommonName:         cn,
		Country:            country,
		Province:           province,
		City:               city,
		Organization:       org,
		OrganizationalUnit: "",
		SubjectAltNames:    sans,
		Expiry:             expireDays,
		Comment:            comment,
	}

	cmd := c.spinner.Start("Requesting certificate...")
	return tea.Batch(cmd, func() tea.Msg {
		cert, err := c.client.RequestSSLCert(context.Background(), req)
		return CertRequestedMsg{Cert: cert, Err: err}
	})
}

// View renders the cert request form.
func (c *CertRequest) View() string {
	var sb strings.Builder
	sb.WriteString(c.form.View())

	if c.spinner.IsActive() {
		sb.WriteString(c.spinner.View())
		sb.WriteString("\n")
	}
	if c.err != "" {
		sb.WriteString(tui.DangerStyle.Render("✗ " + c.err))
		sb.WriteString("\n")
	}
	sb.WriteString(tui.HelpStyle.Render("tab: next field • enter (on last field): submit • esc: back"))
	return sb.String()
}
