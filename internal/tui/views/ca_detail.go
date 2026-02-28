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
caDetailNormal     caDetailMode = iota
caDetailAnalysis               // showing analysis result
caDetailViewCert               // viewing cert PEM content
caDetailChainSel               // selecting chain type before view/export
caDetailExportPath             // entering export file path
caDetailBindUser               // admin: entering username to bind
caDetailBoundList              // admin: listing/unbinding bound users
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
// admin bind/unbind
bindInput   textinput.Model
bindMsg     string
boundUsers  []api.AdminUser
boundTotal  int64
boundPage   int
boundTable  components.Table
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
bi := textinput.New()
bi.Placeholder = "Username to bind"
bi.CharLimit = 128
cols := []components.Column{
{Title: "Username", Width: 22},
{Title: "Display Name", Width: 26},
{Title: "Email", Width: 30},
}
return CADetail{
CA:          ca,
client:      client,
isAdmin:     isAdmin,
spinner:     components.NewSpinner(),
resultVP:    vp,
exportInput: ei,
bindInput:   bi,
boundPage:   1,
boundTable:  components.NewTable(cols, 10),
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
case caDetailBindUser:
switch msg.String() {
case "esc":
c.mode = caDetailNormal
c.bindInput.Blur()
c.bindMsg = ""
return nil
case "enter":
return c.doBindUser()
default:
var cmd tea.Cmd
c.bindInput, cmd = c.bindInput.Update(msg)
return cmd
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
c.mode = caDetailBindUser
c.bindInput.SetValue("")
c.bindInput.Focus()
c.bindMsg = ""
return textinput.Blink
}
case "u":
if c.isAdmin {
c.mode = caDetailBoundList
c.boundPage = 1
return c.loadBoundUsers()
}
}
}
case tea.MouseMsg:
switch c.mode {
case caDetailAnalysis, caDetailViewCert:
var vpCmd tea.Cmd
c.resultVP, vpCmd = c.resultVP.Update(msg)
return vpCmd
case caDetailBoundList:
return c.boundTable.Update(msg)
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
case caBindMsg:
c.spinner.Stop()
if msg.err != "" {
c.bindMsg = "✗ " + msg.err
} else {
c.bindMsg = "✓ Bound successfully"
c.bindInput.SetValue("")
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
}
return c.spinner.Update(msg)
}

// --- message types for bind/unbind ---

type caBindMsg struct{ err string }
type caUnbindMsg struct{ err string }
type caBoundUsersMsg struct {
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
if c.certContent != "" {
if err := writeFile(path, []byte(c.certContent)); err != nil {
c.exportMsg = "Export failed: " + err.Error()
return nil
}
c.exportMsg = "✓ Saved to " + path
c.exportInput.Blur()
c.mode = caDetailViewCert
return nil
}
// Fetch then save.
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
if err != nil {
return inlineAnalysisMsg{err: err.Error()}
}
} else {
certPEM, err = client.GetUserCACert(ctx, uuid, false, false)
if err != nil {
return inlineAnalysisMsg{err: err.Error()}
}
}
analysis, err := client.AnalyzeCert(ctx, certPEM)
if err != nil {
return inlineAnalysisMsg{err: err.Error()}
}
return inlineAnalysisMsg{result: formatCertAnalysis(analysis, vpWidth)}
})
}

func (c *CADetail) doBindUser() tea.Cmd {
username := strings.TrimSpace(c.bindInput.Value())
if username == "" {
c.bindMsg = "Username cannot be empty"
return nil
}
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
case caDetailBindUser:
return c.viewBindUser()
case caDetailBoundList:
return c.viewBoundList()
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

sb.WriteString("\n")
helpKeys := "a: analyze • v: view cert • e: export cert • esc: back"
if c.isAdmin {
helpKeys = "a: analyze • v: view cert • e: export cert • b: bind user • u: bound users • esc: back"
}
sb.WriteString(tui.HelpStyle.Render(helpKeys))
return sb.String()
}

func (c *CADetail) viewAnalysis() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("�� CA Certificate Analysis"))
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
sb.WriteString(tui.HelpStyle.Render("enter: save • esc: back"))
return sb.String()
}

func (c *CADetail) viewBindUser() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("🔐 Bind User to CA"))
sb.WriteString("\n\n")
sb.WriteString(tui.NormalStyle.Render("Enter the username to grant access to this CA:"))
sb.WriteString("\n")
sb.WriteString(tui.InputFocusStyle.Width(c.width - 4).Render(c.bindInput.View()))
sb.WriteString("\n\n")
if c.bindMsg != "" {
if strings.HasPrefix(c.bindMsg, "✓") {
sb.WriteString(tui.SuccessStyle.Render(c.bindMsg))
} else {
sb.WriteString(tui.DangerStyle.Render(c.bindMsg))
}
sb.WriteString("\n")
}
if c.spinner.IsActive() {
sb.WriteString(c.spinner.View())
sb.WriteString("\n")
}
sb.WriteString(tui.HelpStyle.Render("enter: bind • esc: back"))
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
