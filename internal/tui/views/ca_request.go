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

// caReqCAsMsg carries the list of existing CAs for the parent CA selector.
type caReqCAsMsg struct {
cas []api.CACert
err error
}

// CARequest is the form for requesting a new CA certificate.
// It uses a viewport so the form is always scrollable — even in small terminals.
type CARequest struct {
client       *api.Client
fields       []*components.FormField
form         components.Form
viewport     viewport.Model
spinner      components.Spinner
err          string
width        int
height       int
// Parent CA selector state (-1 means "None / Root CA").
availableCAs []api.CACert
parentIdx    int
// Allow Sub-CA toggle (field 1).
allowSubCa bool
// Algorithm selector (field 8).
algoIdx int
}

// NewCARequest creates a new CA request form.
func NewCARequest(client *api.Client) CARequest {
fields := []*components.FormField{
{Label: "Parent CA (↑/↓ to select)", Placeholder: "Loading CAs..."},
{Label: "Allow Sub-CA (↑=true/↓=false/space: toggle)", Placeholder: ""},
{Label: "Common Name (CN)", Placeholder: "e.g. My Root CA"},
{Label: "Country", Placeholder: "e.g. CN"},
{Label: "Province", Placeholder: "e.g. Guangdong"},
{Label: "City", Placeholder: "e.g. Canton"},
{Label: "Organization", Placeholder: "e.g. Acme Corp"},
{Label: "Org Unit", Placeholder: "e.g. IT Department"},
{Label: "Algorithm (↑/↓ to select)", Placeholder: ""},
{Label: "Key Size", Placeholder: "2048/4096 (RSA) • 256/384 (EC) • leave empty for ED25519"},
{Label: "Expire Days", Placeholder: "e.g. 3650"},
{Label: "Comment", Placeholder: "Optional comment"},
}
f := components.NewForm("🔒 Request CA Certificate", fields)
vp := viewport.New(80, 20)
vp.SetContent(f.View())
r := CARequest{
client:    client,
fields:    fields,
form:      f,
viewport:  vp,
spinner:   components.NewSpinner(),
parentIdx: -1, // -1 = None (Root CA)
}
r.form.SetValue(1, "false")
r.form.SetValue(8, certAlgos[0])
return r
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

// Init initializes the form and fetches available parent CAs.
func (c *CARequest) Init() tea.Cmd {
c.form.Reset()
c.err = ""
c.parentIdx = -1
c.availableCAs = nil
c.allowSubCa = false
c.algoIdx = 0
c.form.SetValue(0, "None (Root CA)")
c.form.SetValue(1, "false")
c.form.SetValue(8, certAlgos[0])
c.viewport.GotoTop()
c.refreshViewport()
return tea.Batch(textinput.Blink, c.fetchCAs())
}

// fetchCAs loads all CA certificates (admin view) for the parent selector.
func (c *CARequest) fetchCAs() tea.Cmd {
client := c.client
return func() tea.Msg {
page, err := client.ListAdminCAs(context.Background(), 1, 200)
if err != nil {
return caReqCAsMsg{err: err}
}
return caReqCAsMsg{cas: page.List}
}
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

// parentCaDisplayLabel returns the label for the currently selected parent CA.
func (c *CARequest) parentCaDisplayLabel() string {
if c.parentIdx < 0 {
return "None (Root CA)"
}
ca := c.availableCAs[c.parentIdx]
label := ca.Comment
if label == "" {
label = ca.UUID
}
if len(ca.UUID) > 8 {
label += " (" + ca.UUID[:8] + "...)"
}
return label
}

// selectedParentUUID returns the UUID of the currently selected parent CA, or "" for root.
func (c *CARequest) selectedParentUUID() string {
if c.parentIdx < 0 || c.parentIdx >= len(c.availableCAs) {
return ""
}
return c.availableCAs[c.parentIdx].UUID
}

// Update handles messages.
func (c *CARequest) Update(msg tea.Msg) tea.Cmd {
switch msg := msg.(type) {
case caReqCAsMsg:
if msg.err == nil {
c.availableCAs = msg.cas
}
// Keep parentIdx = -1 (None/Root CA) as the default.
c.form.SetValue(0, c.parentCaDisplayLabel())
c.refreshViewport()
return nil

case tea.KeyMsg:
if c.spinner.IsActive() {
return c.spinner.Update(msg)
}
focused := c.form.FocusedIndex()
// Field 0: Parent CA selector — intercept up/down to cycle.
if focused == 0 {
switch msg.String() {
case "up", "k":
// Cycle backwards: None(-1) → last CA → ... → first CA → None
if c.parentIdx > 0 {
c.parentIdx--
} else if c.parentIdx == 0 {
c.parentIdx = -1
} else {
// Currently None; wrap to last CA.
c.parentIdx = len(c.availableCAs) - 1
}
c.form.SetValue(0, c.parentCaDisplayLabel())
c.refreshViewport()
return nil
case "down", "j":
// Cycle forwards: None(-1) → first CA → ... → last CA → None
if c.parentIdx < len(c.availableCAs)-1 {
c.parentIdx++
} else {
c.parentIdx = -1
}
c.form.SetValue(0, c.parentCaDisplayLabel())
c.refreshViewport()
return nil
case "tab", "shift+tab", "enter":
formCmd := c.form.Update(msg)
c.refreshViewport()
return formCmd
}
// Block free-text editing of the CA selector field.
return nil
}
// Field 1: Allow Sub-CA toggle.
if focused == 1 {
switch msg.String() {
case "up", "k":
c.allowSubCa = true
c.form.SetValue(1, "true")
c.refreshViewport()
return nil
case "down", "j":
c.allowSubCa = false
c.form.SetValue(1, "false")
c.refreshViewport()
return nil
case " ":
c.allowSubCa = !c.allowSubCa
if c.allowSubCa {
c.form.SetValue(1, "true")
} else {
c.form.SetValue(1, "false")
}
c.refreshViewport()
return nil
case "tab", "shift+tab", "enter":
formCmd := c.form.Update(msg)
c.refreshViewport()
return formCmd
}
return nil
}
// Field 8: Algorithm selector.
if focused == 8 {
switch msg.String() {
case "up", "k":
c.algoIdx = (c.algoIdx - 1 + len(certAlgos)) % len(certAlgos)
c.form.SetValue(8, certAlgos[c.algoIdx])
c.refreshViewport()
return nil
case "down", "j":
c.algoIdx = (c.algoIdx + 1) % len(certAlgos)
c.form.SetValue(8, certAlgos[c.algoIdx])
c.refreshViewport()
return nil
case "tab", "shift+tab", "enter":
formCmd := c.form.Update(msg)
c.refreshViewport()
return formCmd
}
return nil
}
switch msg.String() {
case "enter":
if focused == len(c.fields)-1 {
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
parentCaUUID := c.selectedParentUUID()
cn := c.form.Value(2)
country := c.form.Value(3)
province := c.form.Value(4)
city := c.form.Value(5)
org := c.form.Value(6)
ou := c.form.Value(7)
algo := certAlgos[c.algoIdx]
keySizeStr := c.form.Value(9)
expireDaysStr := c.form.Value(10)
comment := c.form.Value(11)

if cn == "" {
c.err = "Common Name is required"
return nil
}

keySize := 0
if keySizeStr != "" {
fmt.Sscanf(keySizeStr, "%d", &keySize)
}

expireDays := 3650
fmt.Sscanf(expireDaysStr, "%d", &expireDays)

if keySize == 0 && (algo == "RSA" || algo == "EC") {
if algo == "RSA" {
keySize = 2048
} else {
keySize = 256
}
}

req := api.RequestCACertRequest{
CaUUID:             parentCaUUID,
AllowSubCa:         c.allowSubCa,
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
helpLine := "tab/↓: next field • shift+tab/↑: prev • enter (last): submit • scroll: mouse wheel"
switch c.form.FocusedIndex() {
case 0:
helpLine = "↑/↓: select parent CA • tab: next field • enter (last): submit"
case 1:
helpLine = "↑: true • ↓: false • space: toggle • tab: next field • enter (last): submit"
case 8:
helpLine = "↑/↓: select algorithm • tab: next field • enter (last): submit"
}
sb.WriteString(tui.HelpStyle.Render(helpLine + scrollInfo))
return sb.String()
}
