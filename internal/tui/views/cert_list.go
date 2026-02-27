package views

import (
"context"
"fmt"
"strings"

tea "github.com/charmbracelet/bubbletea"
"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/components"
)

// CertListLoadedMsg carries SSL cert list data.
type CertListLoadedMsg struct {
Certs []api.SSLCert
Total int64
Err   error
}

// CertList is the SSL certificate list view.
type CertList struct {
client  *api.Client
table   components.Table
certs   []api.SSLCert
total   int64
page    int
spinner components.Spinner
toast   components.Toast
err     string
width   int
height  int
}

// NewCertList creates a new SSL cert list view.
func NewCertList(client *api.Client) CertList {
cols := []components.Column{
{Title: "Comment", Width: 28},
{Title: "Owner", Width: 15},
{Title: "Expires", Width: 12},
{Title: "Days Left", Width: 10},
}
return CertList{
client:  client,
table:   components.NewTable(cols, 15),
page:    1,
spinner: components.NewSpinner(),
}
}

// SetSize updates dimensions.
func (c *CertList) SetSize(width, height int) {
c.width = width
c.height = height
c.table.SetSize(width, height-5)
}

// Init loads certs.
func (c *CertList) Init() tea.Cmd {
cmd := c.spinner.Start("Loading SSL certificates...")
return tea.Batch(cmd, c.load())
}

func (c *CertList) load() tea.Cmd {
return func() tea.Msg {
certs, err := c.client.ListUserSSLCerts(context.Background(), c.page, 20)
if err != nil {
return CertListLoadedMsg{Err: err}
}
return CertListLoadedMsg{Certs: certs.List, Total: certs.Total}
}
}

// SelectedCert returns the currently selected cert.
func (c *CertList) SelectedCert() *api.SSLCert {
idx := c.table.SelectedIndex()
if idx >= 0 && idx < len(c.certs) {
return &c.certs[idx]
}
return nil
}

// Update handles messages.
func (c *CertList) Update(msg tea.Msg) tea.Cmd {
switch msg := msg.(type) {
case CertListLoadedMsg:
c.spinner.Stop()
if msg.Err != nil {
c.err = msg.Err.Error()
return nil
}
c.certs = msg.Certs
c.total = msg.Total
c.table.SetRows(c.buildRows())
return nil

case tea.KeyMsg:
switch msg.String() {
case "r", "f5":
cmd := c.spinner.Start("Refreshing...")
return tea.Batch(cmd, c.load())
case "pgup":
if c.page > 1 {
c.page--
return c.load()
}
case "pgdown":
if int64(c.page*20) < c.total {
c.page++
return c.load()
}
default:
return c.table.Update(msg)
}
}
return c.spinner.Update(msg)
}

func (c *CertList) buildRows() []components.Row {
rows := make([]components.Row, len(c.certs))
for i, cert := range c.certs {
daysLeft := parseDaysLeft(cert.NotAfter)
comment := cert.Comment
if comment == "" {
comment = cert.UUID
}
rows[i] = components.Row{
comment,
cert.Owner,
formatNotAfter(cert.NotAfter),
fmt.Sprintf("%d", daysLeft),
}
}
return rows
}

// View renders the cert list.
func (c *CertList) View() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("📜 SSL Certificates"))
sb.WriteString("\n\n")

if c.spinner.IsActive() {
sb.WriteString(c.spinner.View())
return sb.String()
}
if c.err != "" {
sb.WriteString(tui.DangerStyle.Render("Error: " + c.err))
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("Press r to retry"))
return sb.String()
}

total := fmt.Sprintf("Total: %d | Page: %d", c.total, c.page)
sb.WriteString(tui.MutedStyle.Render(total))
sb.WriteString("\n")
sb.WriteString(c.table.View())
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("enter: details • n: new • d: delete • r: refresh • PgUp/PgDn: page"))

if c.toast.IsVisible() {
sb.WriteString("\n" + c.toast.View())
}
return sb.String()
}
