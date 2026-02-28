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

// CertRequestedMsg is sent when a new SSL cert has been issued.
type CertRequestedMsg struct {
Cert *api.SSLCert
Err  error
}

// certReqCAsMsg carries the list of available CAs for the CA selector.
type certReqCAsMsg struct {
cas []api.CACert
err error
}

// CertRequest is the form for requesting a new SSL certificate.
// It uses a viewport so the form is always scrollable — even in small terminals.
type CertRequest struct {
client        *api.Client
fields        []*components.FormField
form          components.Form
viewport      viewport.Model
spinner       components.Spinner
err           string
width         int
height        int
// CA selector state
availableCAs  []api.CACert
caIdx         int // index into availableCAs; -1 means "none loaded yet"
}

// linesPerField is the number of lines one form field occupies in the viewport.
const linesPerField = 3 // label + input + blank line

// formTitleLines is the number of lines the form title occupies.
const formTitleLines = 2 // title + blank line

// NewCertRequest creates a new cert request form.
func NewCertRequest(client *api.Client) CertRequest {
fields := []*components.FormField{
{Label: "CA (↑/↓ to select)", Placeholder: "Loading available CAs..."},
{Label: "Common Name (CN)", Placeholder: "e.g. example.com"},
{Label: "Country", Placeholder: "e.g. US"},
{Label: "Province", Placeholder: "e.g. California"},
{Label: "City", Placeholder: "e.g. San Francisco"},
{Label: "Organization", Placeholder: "e.g. Acme Corp"},
{Label: "SANs", Placeholder: "Comma-separated: example.com,*.example.com"},
{Label: "Algorithm", Placeholder: "RSA, EC, or ED25519"},
{Label: "Key Size", Placeholder: "2048/4096 (RSA) • 256/384 (EC) • leave empty for ED25519"},
{Label: "Expire Days", Placeholder: "e.g. 365"},
{Label: "Comment", Placeholder: "Optional comment"},
}
// Build a form with NO maxVisible limit; the viewport clips the content.
f := components.NewForm("📜 Request SSL Certificate", fields)
vp := viewport.New(80, 20)
vp.SetContent(f.View())
return CertRequest{
client:   client,
fields:   fields,
form:     f,
viewport: vp,
spinner:  components.NewSpinner(),
caIdx:    -1,
}
}

// SetSize updates dimensions.
func (c *CertRequest) SetSize(width, height int) {
c.width = width
c.height = height
// Reserve bottom lines for spinner/error/help
vpHeight := height - 3
if vpHeight < 3 {
vpHeight = 3
}
c.viewport.Width = width
c.viewport.Height = vpHeight
c.refreshViewport()
}

// Init initializes the form and fetches available CAs.
func (c *CertRequest) Init() tea.Cmd {
c.form.Reset()
c.err = ""
c.caIdx = -1
c.availableCAs = nil
c.viewport.GotoTop()
c.refreshViewport()
return tea.Batch(textinput.Blink, c.fetchCAs())
}

// fetchCAs loads all available CAs (up to 200) for the selector.
func (c *CertRequest) fetchCAs() tea.Cmd {
client := c.client
return func() tea.Msg {
page, err := client.ListUserCAs(context.Background(), 1, 200)
if err != nil {
return certReqCAsMsg{err: err}
}
return certReqCAsMsg{cas: page.List}
}
}

// refreshViewport updates the viewport content and scrolls to show the focused field.
func (c *CertRequest) refreshViewport() {
c.viewport.SetContent(c.form.View())
// Scroll so focused field is visible (centered when possible)
focused := c.form.FocusedIndex()
targetLine := formTitleLines + focused*linesPerField
offset := targetLine - c.viewport.Height/2
if offset < 0 {
offset = 0
}
c.viewport.SetYOffset(offset)
}

// caDisplayLabel returns "comment (short-uuid)" for the currently selected CA.
func (c *CertRequest) caDisplayLabel() string {
if len(c.availableCAs) == 0 {
return ""
}
ca := c.availableCAs[c.caIdx]
label := ca.Comment
if label == "" {
label = ca.UUID
}
if len(ca.UUID) > 8 {
label += " (" + ca.UUID[:8] + "...)"
}
return label
}

// selectedCaUUID returns the UUID of the currently selected CA.
func (c *CertRequest) selectedCaUUID() string {
if c.caIdx < 0 || c.caIdx >= len(c.availableCAs) {
return ""
}
return c.availableCAs[c.caIdx].UUID
}

// Update handles messages.
func (c *CertRequest) Update(msg tea.Msg) tea.Cmd {
switch msg := msg.(type) {
case certReqCAsMsg:
if msg.err == nil && len(msg.cas) > 0 {
c.availableCAs = msg.cas
c.caIdx = 0
c.form.SetValue(0, c.caDisplayLabel())
c.refreshViewport()
} else if msg.err == nil {
c.form.Fields[0].Placeholder = "No CAs available"
}
return nil

case tea.KeyMsg:
if c.spinner.IsActive() {
return c.spinner.Update(msg)
}
// When field 0 (CA selector) is focused, intercept up/down to cycle selections.
if c.form.FocusedIndex() == 0 && len(c.availableCAs) > 0 {
switch msg.String() {
case "up", "k":
if c.caIdx > 0 {
c.caIdx--
} else {
c.caIdx = len(c.availableCAs) - 1
}
c.form.SetValue(0, c.caDisplayLabel())
c.refreshViewport()
return nil
case "down", "j":
if c.caIdx < len(c.availableCAs)-1 {
c.caIdx++
} else {
c.caIdx = 0
}
c.form.SetValue(0, c.caDisplayLabel())
c.refreshViewport()
return nil
}
}
switch msg.String() {
case "enter":
if c.form.FocusedIndex() == len(c.fields)-1 {
return c.submit()
}
// Fall through to form to advance to next field
formCmd := c.form.Update(msg)
c.refreshViewport()
return formCmd
default:
// Don't let the user type into the CA selector field.
if c.form.FocusedIndex() == 0 {
switch msg.String() {
case "tab", "shift+tab":
formCmd := c.form.Update(msg)
c.refreshViewport()
return formCmd
}
return nil
}
formCmd := c.form.Update(msg)
c.refreshViewport()
return formCmd
}

case tea.MouseMsg:
// Forward mouse wheel to viewport for scrolling
var vpCmd tea.Cmd
c.viewport, vpCmd = c.viewport.Update(msg)
return vpCmd

case CertRequestedMsg:
c.spinner.Stop()
if msg.Err != nil {
if isUnauthorized(msg.Err) {
return func() tea.Msg { return SessionExpiredMsg{} }
}
c.err = msg.Err.Error()
}
return nil
}

spinCmd := c.spinner.Update(msg)
return spinCmd
}

func (c *CertRequest) submit() tea.Cmd {
c.err = ""
caUUID := c.selectedCaUUID()
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
c.err = "CN and CA are required"
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

// View renders the cert request form inside a viewport.
func (c *CertRequest) View() string {
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
// Scroll position indicator
pct := c.viewport.ScrollPercent()
scrollInfo := ""
if pct >= 0 && pct <= 1 {
scrollInfo = fmt.Sprintf(" [%.0f%%]", pct*100)
}
helpLine := "tab/↓: next field • shift+tab/↑: prev • enter (last): submit • scroll: mouse wheel"
if c.form.FocusedIndex() == 0 {
helpLine = "↑/↓: select CA • tab: next field • enter (last): submit"
}
sb.WriteString(tui.HelpStyle.Render(helpLine + scrollInfo))
return sb.String()
}
