package views

import (
"context"
"encoding/base64"
"fmt"
"strings"

"github.com/charmbracelet/bubbles/textinput"
"github.com/charmbracelet/bubbles/viewport"
tea "github.com/charmbracelet/bubbletea"
"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/components"
)

// caDetailMode tracks the active sub-mode.
type caDetailMode int

const (
caDetailNormal       caDetailMode = iota
caDetailAnalysis                  // showing analysis result
caDetailViewCert                  // viewing cert PEM content
caDetailChainSel                  // selecting chain type before view/export
caDetailExportPath                // entering export file path
caDetailBindSel                   // admin: selecting unbound user to bind
caDetailBoundList                 // admin: listing/unbinding bound users
caDetailPrivKeyPass               // admin: entering password for private key
caDetailViewPrivKey               // admin: viewing private key PEM content
caDetailExportPriv                // admin: entering export path for private key
)

// caDetailAction is what follows chain selection: view or export.
type caDetailAction int

const (
caDetailActionView   caDetailAction = iota
caDetailActionExport
)

// CADetail shows detailed information about a CA certificate.
type CADetail struct {
CA          *api.CACert
client      *api.Client
isAdmin     bool // use admin API for fetching cert
mode        caDetailMode
spinner     components.Spinner
resultVP    viewport.Model
hasResult   bool
analysisErr string
// cert content
certContent string
chainSelIdx int
chainAction caDetailAction
// export
exportInput textinput.Model
exportMsg   string
// admin bind: unbound user selector
unboundUsers  []api.AdminUser
unboundTotal  int64
unboundPage   int
unboundTable  components.Table
bindMsg       string
// admin bound users list
boundUsers  []api.AdminUser
boundTotal  int64
boundPage   int
boundTable  components.Table
// private key
passInput   textinput.Model
privKeyMode caDetailAction
privContent string
privExport  textinput.Model
privMsg     string
width       int
height      int
}

// NewCADetail creates a new CA detail view.
// Pass isAdmin=true when used inside the Admin view (uses admin API endpoints).
func NewCADetail(ca *api.CACert, client *api.Client, isAdmin bool) CADetail {
vp := viewport.New(80, 20)
ei := textinput.New()
ei.Placeholder = "e.g. /home/user/ca.pem"
ei.CharLimit = 512
pi := textinput.New()
pi.Placeholder = "Enter your login password"
pi.EchoMode = textinput.EchoPassword
pi.CharLimit = 256
pe := textinput.New()
pe.Placeholder = "e.g. /home/user/ca.key"
pe.CharLimit = 512
ubCols := []components.Column{
{Title: "Username", Width: 22},
{Title: "Display Name", Width: 26},
{Title: "Email", Width: 30},
}
bCols := []components.Column{
{Title: "Username", Width: 22},
{Title: "Display Name", Width: 26},
{Title: "Email", Width: 30},
}
return CADetail{
CA:           ca,
client:       client,
isAdmin:      isAdmin,
spinner:      components.NewSpinner(),
resultVP:     vp,
exportInput:  ei,
passInput:    pi,
privExport:   pe,
unboundPage:  1,
unboundTable: components.NewTable(ubCols, 10),
boundPage:    1,
boundTable:   components.NewTable(bCols, 10),
}
}

// SetSize updates dimensions.
func (c *CADetail) SetSize(width, height int) {
c.width = width
c.height = height
c.resultVP.Width = width
vpH := height - 5
if vpH < 4 {
vpH = 4
}
c.resultVP.Height = vpH
c.unboundTable.SetSize(width, height-8)
c.boundTable.SetSize(width, height-8)
}

// Init does nothing for this static view.
func (c *CADetail) Init() tea.Cmd { return nil }

// IsAnalysisMode returns true when any sub-mode is active (prevents outer esc handling).
func (c *CADetail) IsAnalysisMode() bool { return c.mode != caDetailNormal }

// Update handles messages.
func (c *CADetail) Update(msg tea.Msg) tea.Cmd {
switch msg := msg.(type) {
case tea.KeyMsg:
if c.spinner.IsActive() {
return c.spinner.Update(msg)
}
switch c.mode {
case caDetailAnalysis:
switch msg.String() {
case "esc":
c.mode = caDetailNormal
c.hasResult = false
c.analysisErr = ""
return nil
case "up", "k", "down", "j", "pgup", "pgdown":
var vpCmd tea.Cmd
c.resultVP, vpCmd = c.resultVP.Update(msg)
return vpCmd
}
case caDetailViewCert:
switch msg.String() {
case "esc":
c.mode = caDetailNormal
c.certContent = ""
return nil
case "e":
c.mode = caDetailExportPath
c.exportInput.SetValue("")
c.exportInput.Focus()
c.exportMsg = ""
return textinput.Blink
case "up", "k", "down", "j", "pgup", "pgdown":
var vpCmd tea.Cmd
c.resultVP, vpCmd = c.resultVP.Update(msg)
return vpCmd
}
case caDetailChainSel:
switch msg.String() {
case "esc":
c.mode = caDetailNormal
return nil
case "up", "k":
if c.chainSelIdx > 0 {
c.chainSelIdx--
}
case "down", "j":
if c.chainSelIdx < len(certChainOptions)-1 {
c.chainSelIdx++
}
case "enter":
opt := certChainOptions[c.chainSelIdx]
if c.chainAction == caDetailActionView {
return c.fetchAndViewCert(opt.chain, opt.needRoot)
}
c.mode = caDetailExportPath
c.exportInput.SetValue("")
c.exportInput.Focus()
c.exportMsg = ""
return textinput.Blink
}
case caDetailExportPath:
switch msg.String() {
case "esc":
c.mode = caDetailViewCert
c.exportInput.Blur()
return nil
case "enter":
return c.doExportCA()
default:
var cmd tea.Cmd
c.exportInput, cmd = c.exportInput.Update(msg)
return cmd
}
case caDetailBindSel:
switch msg.String() {
case "esc":
c.mode = caDetailNormal
c.bindMsg = ""
return nil
case "enter":
idx := c.unboundTable.SelectedIndex()
if idx >= 0 && idx < len(c.unboundUsers) {
return c.doBindUser(c.unboundUsers[idx].Username)
}
case "[":
if c.unboundPage > 1 {
c.unboundPage--
return c.loadUnboundUsers()
}
case "]":
if int64(c.unboundPage*20) < c.unboundTotal {
c.unboundPage++
return c.loadUnboundUsers()
}
default:
return c.unboundTable.Update(msg)
}
case caDetailBoundList:
switch msg.String() {
case "esc":
c.mode = caDetailNormal
return nil
case "d", "delete":
idx := c.boundTable.SelectedIndex()
if idx >= 0 && idx < len(c.boundUsers) {
return c.doUnbindUser(c.boundUsers[idx].Username)
}
case "[":
if c.boundPage > 1 {
c.boundPage--
return c.loadBoundUsers()
}
case "]":
if int64(c.boundPage*20) < c.boundTotal {
c.boundPage++
return c.loadBoundUsers()
}
default:
return c.boundTable.Update(msg)
}
case caDetailPrivKeyPass:
switch msg.String() {
case "esc":
c.mode = caDetailNormal
c.passInput.Blur()
c.passInput.SetValue("")
return nil
case "enter":
return c.fetchPrivKey()
default:
var cmd tea.Cmd
c.passInput, cmd = c.passInput.Update(msg)
return cmd
}
case caDetailViewPrivKey:
switch msg.String() {
case "esc":
c.mode = caDetailNormal
c.privContent = ""
c.privMsg = ""
return nil
case "e":
c.mode = caDetailExportPriv
c.privExport.SetValue("")
c.privExport.Focus()
c.privMsg = ""
return textinput.Blink
case "up", "k", "down", "j", "pgup", "pgdown":
var vpCmd tea.Cmd
c.resultVP, vpCmd = c.resultVP.Update(msg)
return vpCmd
}
case caDetailExportPriv:
switch msg.String() {
case "esc":
c.mode = caDetailViewPrivKey
c.privExport.Blur()
return nil
case "enter":
return c.doExportPrivKey()
default:
var cmd tea.Cmd
c.privExport, cmd = c.privExport.Update(msg)
return cmd
}
case caDetailNormal:
switch msg.String() {
case "a":
return c.startAnalysis()
case "v":
c.chainAction = caDetailActionView
c.chainSelIdx = 0
c.mode = caDetailChainSel
case "e":
c.chainAction = caDetailActionExport
c.chainSelIdx = 0
c.mode = caDetailChainSel
case "b":
if c.isAdmin {
c.mode = caDetailBindSel
c.unboundPage = 1
c.bindMsg = ""
return c.loadUnboundUsers()
}
case "u":
if c.isAdmin {
c.mode = caDetailBoundList
c.boundPage = 1
return c.loadBoundUsers()
}
case "k":
if c.isAdmin {
c.privKeyMode = caDetailActionView
c.passInput.SetValue("")
c.passInput.Focus()
c.privMsg = ""
c.mode = caDetailPrivKeyPass
return textinput.Blink
}
case "K":
if c.isAdmin {
c.privKeyMode = caDetailActionExport
c.passInput.SetValue("")
c.passInput.Focus()
c.privMsg = ""
c.mode = caDetailPrivKeyPass
return textinput.Blink
}
}
}
case tea.MouseMsg:
switch c.mode {
case caDetailAnalysis, caDetailViewCert, caDetailViewPrivKey:
var vpCmd tea.Cmd
c.resultVP, vpCmd = c.resultVP.Update(msg)
return vpCmd
case caDetailBoundList:
return c.boundTable.Update(msg)
case caDetailBindSel:
return c.unboundTable.Update(msg)
}
case inlineAnalysisMsg:
c.spinner.Stop()
c.mode = caDetailAnalysis
c.analysisErr = msg.err
if msg.result != "" {
c.hasResult = true
c.resultVP.SetContent(msg.result)
c.resultVP.GotoTop()
}
return nil
case certContentMsg:
c.spinner.Stop()
if msg.err != "" {
c.analysisErr = msg.err
c.mode = caDetailAnalysis
return nil
}
c.certContent = msg.content
c.mode = caDetailViewCert
c.resultVP.SetContent(msg.content)
c.resultVP.GotoTop()
return nil
case certExportedMsg:
c.spinner.Stop()
c.exportMsg = "✓ Saved to " + msg.path
c.exportInput.Blur()
c.mode = caDetailViewCert
return nil
case certPrivKeyMsg:
c.spinner.Stop()
if msg.err != "" {
c.privMsg = "✗ " + msg.err
c.mode = caDetailNormal
return nil
}
c.privContent = msg.content
if c.privKeyMode == caDetailActionExport {
c.mode = caDetailExportPriv
c.privExport.SetValue("")
c.privExport.Focus()
c.privMsg = ""
return textinput.Blink
}
c.mode = caDetailViewPrivKey
c.resultVP.SetContent(msg.content)
c.resultVP.GotoTop()
return nil
case certPrivExportedMsg:
c.spinner.Stop()
c.privMsg = "✓ Saved to " + msg.path
c.privExport.Blur()
c.mode = caDetailViewPrivKey
return nil
case caBindMsg:
c.spinner.Stop()
if msg.err != "" {
c.bindMsg = "✗ " + msg.err
} else {
c.bindMsg = "✓ Bound successfully"
// Refresh unbound list.
return c.loadUnboundUsers()
}
return nil
case caUnbindMsg:
c.spinner.Stop()
if msg.err != "" {
c.bindMsg = "✗ " + msg.err
}
return c.loadBoundUsers()
case caBoundUsersMsg:
c.spinner.Stop()
if msg.err != "" {
c.bindMsg = "✗ " + msg.err
return nil
}
c.boundUsers = msg.users
c.boundTotal = msg.total
rows := make([]components.Row, len(msg.users))
for i, u := range msg.users {
rows[i] = components.Row{u.Username, u.DisplayName, u.Email}
}
c.boundTable.SetRows(rows)
return nil
case caUnboundUsersMsg:
c.spinner.Stop()
if msg.err != "" {
c.bindMsg = "✗ " + msg.err
return nil
}
c.unboundUsers = msg.users
c.unboundTotal = msg.total
rows := make([]components.Row, len(msg.users))
for i, u := range msg.users {
rows[i] = components.Row{u.Username, u.DisplayName, u.Email}
}
c.unboundTable.SetRows(rows)
return nil
}
return c.spinner.Update(msg)
}

// --- message types ---

type caBindMsg struct{ err string }
type caUnbindMsg struct{ err string }
type caBoundUsersMsg struct {
users []api.AdminUser
total int64
err   string
}
type caUnboundUsersMsg struct {
users []api.AdminUser
total int64
err   string
}

// --- helper commands ---

func (c *CADetail) fetchAndViewCert(chain, needRoot bool) tea.Cmd {
uuid := c.CA.UUID
client := c.client
isAdmin := c.isAdmin
spinCmd := c.spinner.Start("Fetching certificate...")
return tea.Batch(spinCmd, func() tea.Msg {
ctx := context.Background()
var encoded string
var err error
if isAdmin {
encoded, err = client.GetAdminCACert(ctx, uuid, chain, needRoot)
} else {
encoded, err = client.GetUserCACert(ctx, uuid, chain, needRoot)
}
if err != nil {
return certContentMsg{err: err.Error()}
}
decoded, decErr := base64.StdEncoding.DecodeString(encoded)
if decErr != nil {
return certContentMsg{err: "decode error: " + decErr.Error()}
}
return certContentMsg{content: string(decoded)}
})
}

func (c *CADetail) doExportCA() tea.Cmd {
path := strings.TrimSpace(c.exportInput.Value())
if path == "" {
c.exportMsg = "Path cannot be empty"
return nil
}
c.exportInput.Blur()
content := c.certContent
if content != "" {
spinCmd := c.spinner.Start("Saving...")
return tea.Batch(spinCmd, func() tea.Msg {
if err := writeFile(path, []byte(content)); err != nil {
return certContentMsg{err: "export failed: " + err.Error()}
}
return certExportedMsg{path: path}
})
}
opt := certChainOptions[c.chainSelIdx]
uuid := c.CA.UUID
client := c.client
isAdmin := c.isAdmin
spinCmd := c.spinner.Start("Exporting...")
return tea.Batch(spinCmd, func() tea.Msg {
ctx := context.Background()
var encoded string
var err error
if isAdmin {
encoded, err = client.GetAdminCACert(ctx, uuid, opt.chain, opt.needRoot)
} else {
encoded, err = client.GetUserCACert(ctx, uuid, opt.chain, opt.needRoot)
}
if err != nil {
return certContentMsg{err: err.Error()}
}
decoded, decErr := base64.StdEncoding.DecodeString(encoded)
if decErr != nil {
return certContentMsg{err: "decode error: " + decErr.Error()}
}
if writeErr := writeFile(path, decoded); writeErr != nil {
return certContentMsg{err: writeErr.Error()}
}
return certExportedMsg{path: path}
})
}

func (c *CADetail) startAnalysis() tea.Cmd {
uuid := c.CA.UUID
client := c.client
isAdmin := c.isAdmin
vpWidth := c.resultVP.Width
spinCmd := c.spinner.Start("Analyzing...")
return tea.Batch(spinCmd, func() tea.Msg {
ctx := context.Background()
var certPEM string
var err error
if isAdmin {
certPEM, err = client.GetAdminCACert(ctx, uuid, false, false)
} else {
certPEM, err = client.GetUserCACert(ctx, uuid, false, false)
}
if err != nil {
return inlineAnalysisMsg{err: err.Error()}
}
analysis, err := client.AnalyzeCert(ctx, certPEM)
if err != nil {
return inlineAnalysisMsg{err: err.Error()}
}
return inlineAnalysisMsg{result: formatCertAnalysis(analysis, vpWidth)}
})
}

func (c *CADetail) fetchPrivKey() tea.Cmd {
password := c.passInput.Value()
uuid := c.CA.UUID
client := c.client
c.passInput.Blur()
c.passInput.SetValue("")
spinCmd := c.spinner.Start("Fetching private key...")
return tea.Batch(spinCmd, func() tea.Msg {
resp, err := client.GetAdminCAPrivKey(context.Background(), uuid, password)
if err != nil {
return certPrivKeyMsg{err: err.Error()}
}
decoded, decErr := base64.StdEncoding.DecodeString(resp)
if decErr != nil {
return certPrivKeyMsg{err: "decode error: " + decErr.Error()}
}
return certPrivKeyMsg{content: string(decoded)}
})
}

func (c *CADetail) doExportPrivKey() tea.Cmd {
path := strings.TrimSpace(c.privExport.Value())
if path == "" {
c.privMsg = "Path cannot be empty"
return nil
}
c.privExport.Blur()
content := c.privContent
spinCmd := c.spinner.Start("Saving private key...")
return tea.Batch(spinCmd, func() tea.Msg {
if err := writeFile(path, []byte(content)); err != nil {
return certPrivKeyMsg{err: "export failed: " + err.Error()}
}
return certPrivExportedMsg{path: path}
})
}

func (c *CADetail) doBindUser(username string) tea.Cmd {
uuid := c.CA.UUID
client := c.client
spinCmd := c.spinner.Start("Binding user...")
return tea.Batch(spinCmd, func() tea.Msg {
err := client.BindUsersToCA(context.Background(), uuid, []string{username})
if err != nil {
return caBindMsg{err: err.Error()}
}
return caBindMsg{}
})
}

func (c *CADetail) doUnbindUser(username string) tea.Cmd {
uuid := c.CA.UUID
client := c.client
spinCmd := c.spinner.Start("Unbinding user...")
return tea.Batch(spinCmd, func() tea.Msg {
err := client.UnbindUsersFromCA(context.Background(), uuid, []string{username})
if err != nil {
return caUnbindMsg{err: err.Error()}
}
return caUnbindMsg{}
})
}

func (c *CADetail) loadBoundUsers() tea.Cmd {
uuid := c.CA.UUID
page := c.boundPage
client := c.client
spinCmd := c.spinner.Start("Loading bound users...")
return tea.Batch(spinCmd, func() tea.Msg {
result, err := client.GetBoundUsers(context.Background(), uuid, page, 20)
if err != nil {
return caBoundUsersMsg{err: err.Error()}
}
return caBoundUsersMsg{users: result.List, total: result.Total}
})
}

func (c *CADetail) loadUnboundUsers() tea.Cmd {
uuid := c.CA.UUID
page := c.unboundPage
client := c.client
spinCmd := c.spinner.Start("Loading unbound users...")
return tea.Batch(spinCmd, func() tea.Msg {
result, err := client.GetUnboundUsers(context.Background(), uuid, page, 20)
if err != nil {
return caUnboundUsersMsg{err: err.Error()}
}
return caUnboundUsersMsg{users: result.List, total: result.Total}
})
}

// View renders the CA detail.
func (c *CADetail) View() string {
if c.CA == nil {
return tui.MutedStyle.Render("No CA selected.")
}

switch c.mode {
case caDetailAnalysis:
return c.viewAnalysis()
case caDetailViewCert:
return c.viewCertContent()
case caDetailChainSel:
return c.viewChainSelector()
case caDetailExportPath:
return c.viewExportPath()
case caDetailBindSel:
return c.viewBindSel()
case caDetailBoundList:
return c.viewBoundList()
case caDetailPrivKeyPass:
return c.viewPrivKeyPass()
case caDetailViewPrivKey:
return c.viewPrivKey()
case caDetailExportPriv:
return c.viewExportPriv()
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
{"Type", tui.CATypeStyle(ca.CAType()).Render(ca.CAType())},
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

if c.privMsg != "" {
sb.WriteString("\n")
sb.WriteString(tui.DangerStyle.Render(c.privMsg))
sb.WriteString("\n")
}

sb.WriteString("\n")
helpKeys := "a: analyze • v: view cert • e: export cert • esc: back"
if c.isAdmin {
helpKeys = "a: analyze • v: view cert • e: export cert • k: view privkey • K: export privkey • b: bind user • u: bound users • esc: back"
}
sb.WriteString(tui.HelpStyle.Render(helpKeys))
return sb.String()
}

func (c *CADetail) viewAnalysis() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("🔐 CA Certificate Analysis"))
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

func (c *CADetail) viewCertContent() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("🔐 CA Certificate Content"))
sb.WriteString("\n\n")

if c.spinner.IsActive() {
sb.WriteString(c.spinner.View())
sb.WriteString("\n")
} else {
sb.WriteString(c.resultVP.View())
if c.resultVP.TotalLineCount() > c.resultVP.Height {
pct := int(c.resultVP.ScrollPercent() * 100)
sb.WriteString(tui.MutedStyle.Render(fmt.Sprintf(" %d%%", pct)))
}
sb.WriteString("\n")
}

if c.exportMsg != "" {
sb.WriteString(tui.SuccessStyle.Render(c.exportMsg))
sb.WriteString("\n")
}

sb.WriteString(tui.HelpStyle.Render("↑/↓: scroll • e: export to file • esc: back"))
return sb.String()
}

func (c *CADetail) viewChainSelector() string {
var sb strings.Builder
actionLabel := "View"
if c.chainAction == caDetailActionExport {
actionLabel = "Export"
}
sb.WriteString(tui.TitleStyle.Render(fmt.Sprintf("🔐 %s Certificate — Select Type", actionLabel)))
sb.WriteString("\n\n")

for i, opt := range certChainOptions {
if i == c.chainSelIdx {
sb.WriteString(tui.SelectedStyle.Render("▶ " + opt.label))
} else {
sb.WriteString(tui.NormalStyle.Render("  " + opt.label))
}
sb.WriteString("\n")
}

sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("↑/↓: select • enter: confirm • esc: back"))
return sb.String()
}

func (c *CADetail) viewExportPath() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("🔐 Export CA Certificate"))
sb.WriteString("\n\n")
sb.WriteString(tui.NormalStyle.Render("Enter the file path to save the certificate:"))
sb.WriteString("\n")
sb.WriteString(tui.InputFocusStyle.Width(c.width - 4).Render(c.exportInput.View()))
sb.WriteString("\n\n")
if c.exportMsg != "" {
if strings.HasPrefix(c.exportMsg, "✓") {
sb.WriteString(tui.SuccessStyle.Render(c.exportMsg))
} else {
sb.WriteString(tui.DangerStyle.Render(c.exportMsg))
}
sb.WriteString("\n")
}
if c.spinner.IsActive() {
sb.WriteString(c.spinner.View())
sb.WriteString("\n")
}
sb.WriteString(tui.HelpStyle.Render("enter: save • esc: back"))
return sb.String()
}

func (c *CADetail) viewBindSel() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("🔐 Bind User to CA — Select User"))
sb.WriteString("\n\n")

if c.spinner.IsActive() {
sb.WriteString(c.spinner.View())
return sb.String()
}

if c.bindMsg != "" {
if strings.HasPrefix(c.bindMsg, "✓") {
sb.WriteString(tui.SuccessStyle.Render(c.bindMsg))
} else {
sb.WriteString(tui.DangerStyle.Render(c.bindMsg))
}
sb.WriteString("\n")
}

total := fmt.Sprintf("Total: %d | Page: %d", c.unboundTotal, c.unboundPage)
sb.WriteString(tui.MutedStyle.Render(total))
sb.WriteString("\n")
sb.WriteString(c.unboundTable.View())
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("enter: bind selected • [/]: prev/next page • esc: back"))
return sb.String()
}

func (c *CADetail) viewBoundList() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("🔐 Bound Users"))
sb.WriteString("\n\n")

if c.spinner.IsActive() {
sb.WriteString(c.spinner.View())
return sb.String()
}

if c.bindMsg != "" {
sb.WriteString(tui.DangerStyle.Render(c.bindMsg))
sb.WriteString("\n")
}

total := fmt.Sprintf("Total: %d | Page: %d", c.boundTotal, c.boundPage)
sb.WriteString(tui.MutedStyle.Render(total))
sb.WriteString("\n")
sb.WriteString(c.boundTable.View())
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("d: unbind selected • [/]: prev/next page • esc: back"))
return sb.String()
}

func (c *CADetail) viewPrivKeyPass() string {
var sb strings.Builder
action := "View"
if c.privKeyMode == caDetailActionExport {
action = "Export"
}
sb.WriteString(tui.TitleStyle.Render(fmt.Sprintf("🔑 %s CA Private Key — Authentication", action)))
sb.WriteString("\n\n")
sb.WriteString(tui.NormalStyle.Render("Enter your account password to decrypt the private key:"))
sb.WriteString("\n")
sb.WriteString(tui.InputFocusStyle.Width(c.width - 4).Render(c.passInput.View()))
sb.WriteString("\n\n")
if c.spinner.IsActive() {
sb.WriteString(c.spinner.View())
sb.WriteString("\n")
}
sb.WriteString(tui.HelpStyle.Render("enter: confirm • esc: back"))
return sb.String()
}

func (c *CADetail) viewPrivKey() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("🔑 CA Certificate Private Key"))
sb.WriteString("\n\n")

if c.spinner.IsActive() {
sb.WriteString(c.spinner.View())
sb.WriteString("\n")
} else {
sb.WriteString(c.resultVP.View())
if c.resultVP.TotalLineCount() > c.resultVP.Height {
pct := int(c.resultVP.ScrollPercent() * 100)
sb.WriteString(tui.MutedStyle.Render(fmt.Sprintf(" %d%%", pct)))
}
sb.WriteString("\n")
}

if c.privMsg != "" {
sb.WriteString(tui.SuccessStyle.Render(c.privMsg))
sb.WriteString("\n")
}

sb.WriteString(tui.HelpStyle.Render("↑/↓: scroll • e: export to file • esc: back"))
return sb.String()
}

func (c *CADetail) viewExportPriv() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("🔑 Export CA Private Key"))
sb.WriteString("\n\n")
sb.WriteString(tui.NormalStyle.Render("Enter the file path to save the private key:"))
sb.WriteString("\n")
sb.WriteString(tui.InputFocusStyle.Width(c.width - 4).Render(c.privExport.View()))
sb.WriteString("\n\n")
if c.privMsg != "" {
if strings.HasPrefix(c.privMsg, "✓") {
sb.WriteString(tui.SuccessStyle.Render(c.privMsg))
} else {
sb.WriteString(tui.DangerStyle.Render(c.privMsg))
}
sb.WriteString("\n")
}
if c.spinner.IsActive() {
sb.WriteString(c.spinner.View())
sb.WriteString("\n")
}
sb.WriteString(tui.HelpStyle.Render("enter: save • esc: back"))
return sb.String()
}
