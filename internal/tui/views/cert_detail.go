package views

import (
"context"
"encoding/base64"
"fmt"
"os"
"path/filepath"
"strings"

"github.com/charmbracelet/bubbles/textinput"
"github.com/charmbracelet/bubbles/viewport"
tea "github.com/charmbracelet/bubbletea"
"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/components"
)

// certDetailMode tracks the active sub-mode.
type certDetailMode int

const (
certDetailNormal      certDetailMode = iota
certDetailAnalysis                   // showing analysis result
certDetailViewCert                   // viewing cert PEM content
certDetailChainSel                   // selecting chain type before view/export
certDetailExportPath                 // entering export file path
certDetailPrivKeyPass                // entering password for private key
certDetailViewPrivKey                // viewing private key PEM content
certDetailExportPriv                 // entering export path for private key
)

// chainTypeOption describes a certificate chain option.
type chainTypeOption struct {
label    string
chain    bool
needRoot bool
}

var certChainOptions = []chainTypeOption{
{"Only This Certificate", false, false},
{"Full Chain (with Root CA)", true, true},
{"Chain (without Root CA)", true, false},
}

// certDetailAction is what we do after chain selection: view or export.
type certDetailAction int

const (
certDetailActionView   certDetailAction = iota
certDetailActionExport
)

// CertDetail shows detailed information about an SSL certificate.
type CertDetail struct {
Cert         *api.SSLCert
client       *api.Client
mode         certDetailMode
spinner      components.Spinner
resultVP     viewport.Model
hasResult    bool
analysisErr  string
// cert content view
certContent string // decoded PEM content
chainSelIdx int
chainAction certDetailAction
// export path input
exportInput components.PathInput
exportMsg   string
// private key
passInput   textinput.Model
privKeyMode certDetailAction // view or export
privContent string
privExport  components.PathInput
privMsg     string
width       int
height      int
}

// NewCertDetail creates a new SSL cert detail view.
func NewCertDetail(cert *api.SSLCert, client *api.Client) CertDetail {
vp := viewport.New(80, 20)
ei := components.NewPathInput("e.g. /home/user/cert.pem", 512)
pi := textinput.New()
pi.Placeholder = "Enter your login password"
pi.EchoMode = textinput.EchoPassword
pi.CharLimit = 256
pe := components.NewPathInput("e.g. /home/user/key.pem", 512)
return CertDetail{
Cert:       cert,
client:     client,
spinner:    components.NewSpinner(),
resultVP:   vp,
exportInput: ei,
passInput:  pi,
privExport: pe,
}
}

// SetSize updates dimensions.
func (c *CertDetail) SetSize(width, height int) {
c.width = width
c.height = height
c.resultVP.Width = width
vpH := height - 5
if vpH < 4 {
vpH = 4
}
c.resultVP.Height = vpH
}

// Init does nothing.
func (c *CertDetail) Init() tea.Cmd { return nil }

// IsAnalysisMode returns true when any sub-mode is active (prevents outer esc handling).
func (c *CertDetail) IsAnalysisMode() bool { return c.mode != certDetailNormal }

// Update handles messages.
func (c *CertDetail) Update(msg tea.Msg) tea.Cmd {
switch msg := msg.(type) {
case tea.KeyMsg:
if c.spinner.IsActive() {
return c.spinner.Update(msg)
}
switch c.mode {
case certDetailAnalysis:
switch msg.String() {
case "esc":
c.mode = certDetailNormal
c.hasResult = false
c.analysisErr = ""
return nil
case "up", "k", "down", "j", "pgup", "pgdown":
var vpCmd tea.Cmd
c.resultVP, vpCmd = c.resultVP.Update(msg)
return vpCmd
}
case certDetailViewCert:
switch msg.String() {
case "esc":
c.mode = certDetailNormal
c.certContent = ""
return nil
case "e":
c.mode = certDetailExportPath
c.exportInput.SetValue("")
c.exportInput.Focus()
c.exportMsg = ""
return textinput.Blink
case "up", "k", "down", "j", "pgup", "pgdown":
var vpCmd tea.Cmd
c.resultVP, vpCmd = c.resultVP.Update(msg)
return vpCmd
}
case certDetailChainSel:
switch msg.String() {
case "esc":
c.mode = certDetailNormal
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
if c.chainAction == certDetailActionView {
return c.fetchAndViewCert(opt.chain, opt.needRoot)
}
c.mode = certDetailExportPath
c.exportInput.SetValue("")
c.exportInput.Focus()
c.exportMsg = ""
return textinput.Blink
}
case certDetailExportPath:
switch msg.String() {
case "esc":
c.mode = certDetailViewCert
c.exportInput.Blur()
return nil
case "enter":
return c.doExport()
default:
cmd := c.exportInput.Update(msg)
return cmd
}
case certDetailPrivKeyPass:
switch msg.String() {
case "esc":
c.mode = certDetailNormal
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
case certDetailViewPrivKey:
switch msg.String() {
case "esc":
c.mode = certDetailNormal
c.privContent = ""
c.privMsg = ""
return nil
case "e":
c.mode = certDetailExportPriv
c.privExport.SetValue("")
c.privExport.Focus()
c.privMsg = ""
return textinput.Blink
case "up", "k", "down", "j", "pgup", "pgdown":
var vpCmd tea.Cmd
c.resultVP, vpCmd = c.resultVP.Update(msg)
return vpCmd
}
case certDetailExportPriv:
switch msg.String() {
case "esc":
c.mode = certDetailViewPrivKey
c.privExport.Blur()
return nil
case "enter":
return c.doExportPrivKey()
default:
cmd := c.privExport.Update(msg)
return cmd
}
case certDetailNormal:
switch msg.String() {
case "a":
return c.startAnalysis()
case "v":
c.chainAction = certDetailActionView
c.chainSelIdx = 0
c.mode = certDetailChainSel
case "e":
c.chainAction = certDetailActionExport
c.chainSelIdx = 0
c.mode = certDetailChainSel
case "k":
c.privKeyMode = certDetailActionView
c.passInput.SetValue("")
c.passInput.Focus()
c.privMsg = ""
c.mode = certDetailPrivKeyPass
return textinput.Blink
case "K":
c.privKeyMode = certDetailActionExport
c.passInput.SetValue("")
c.passInput.Focus()
c.privMsg = ""
c.mode = certDetailPrivKeyPass
return textinput.Blink
}
}
case tea.MouseMsg:
if c.mode == certDetailAnalysis || c.mode == certDetailViewCert || c.mode == certDetailViewPrivKey {
var vpCmd tea.Cmd
c.resultVP, vpCmd = c.resultVP.Update(msg)
return vpCmd
}
case inlineAnalysisMsg:
c.spinner.Stop()
c.mode = certDetailAnalysis
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
c.mode = certDetailAnalysis
return nil
}
c.certContent = msg.content
c.mode = certDetailViewCert
c.resultVP.SetContent(msg.content)
c.resultVP.GotoTop()
return nil
case certPrivKeyMsg:
c.spinner.Stop()
if msg.err != "" {
c.privMsg = "✗ " + msg.err
c.mode = certDetailNormal
return nil
}
c.privContent = msg.content
if c.privKeyMode == certDetailActionExport {
c.mode = certDetailExportPriv
c.privExport.SetValue("")
c.privExport.Focus()
c.privMsg = ""
return textinput.Blink
}
c.mode = certDetailViewPrivKey
c.resultVP.SetContent(msg.content)
c.resultVP.GotoTop()
return nil
case certExportedMsg:
c.spinner.Stop()
c.exportMsg = "✓ Saved to " + msg.path
c.exportInput.Blur()
c.mode = certDetailViewCert
return nil
case certPrivExportedMsg:
c.spinner.Stop()
c.privMsg = "✓ Saved to " + msg.path
c.privExport.Blur()
c.mode = certDetailViewPrivKey
return nil
}
return c.spinner.Update(msg)
}

// certContentMsg carries decoded PEM content.
type certContentMsg struct {
content string
err     string
}

// certPrivKeyMsg carries a decoded private key PEM.
type certPrivKeyMsg struct {
content string
err     string
}

// certExportedMsg signals a successful cert export.
type certExportedMsg struct{ path string }

// certPrivExportedMsg signals a successful private key export.
type certPrivExportedMsg struct{ path string }

func (c *CertDetail) fetchAndViewCert(chain, needRoot bool) tea.Cmd {
uuid := c.Cert.UUID
client := c.client
spinCmd := c.spinner.Start("Fetching certificate...")
return tea.Batch(spinCmd, func() tea.Msg {
ctx := context.Background()
encoded, err := client.GetUserSSLCert(ctx, uuid, chain, needRoot)
if err != nil {
return certContentMsg{err: err.Error()}
}
decoded, err := base64.StdEncoding.DecodeString(encoded)
if err != nil {
return certContentMsg{err: "decode error: " + err.Error()}
}
return certContentMsg{content: string(decoded)}
})
}

func (c *CertDetail) fetchPrivKey() tea.Cmd {
password := c.passInput.Value()
uuid := c.Cert.UUID
client := c.client
c.passInput.Blur()
c.passInput.SetValue("")
spinCmd := c.spinner.Start("Fetching private key...")
return tea.Batch(spinCmd, func() tea.Msg {
resp, err := client.GetUserSSLPrivKey(context.Background(), uuid, password)
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

func (c *CertDetail) doExport() tea.Cmd {
path := strings.TrimSpace(c.exportInput.Value())
if path == "" {
c.exportMsg = "Path cannot be empty"
return nil
}
c.exportInput.Blur()
content := c.certContent
if content != "" {
// Content already fetched — write asynchronously to avoid blocking UI.
spinCmd := c.spinner.Start("Saving...")
return tea.Batch(spinCmd, func() tea.Msg {
if err := writeFile(path, []byte(content)); err != nil {
return certContentMsg{err: "export failed: " + err.Error()}
}
return certExportedMsg{path: path}
})
}
// Need to fetch then save.
opt := certChainOptions[c.chainSelIdx]
uuid := c.Cert.UUID
client := c.client
spinCmd := c.spinner.Start("Exporting...")
return tea.Batch(spinCmd, func() tea.Msg {
ctx := context.Background()
encoded, err := client.GetUserSSLCert(ctx, uuid, opt.chain, opt.needRoot)
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

func (c *CertDetail) doExportPrivKey() tea.Cmd {
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

// writeFile creates parent directories and writes data to path.
func writeFile(path string, data []byte) error {
if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
return err
}
return os.WriteFile(path, data, 0600)
}

func (c *CertDetail) startAnalysis() tea.Cmd {
uuid := c.Cert.UUID
client := c.client
vpWidth := c.resultVP.Width
spinCmd := c.spinner.Start("Analyzing...")
return tea.Batch(spinCmd, func() tea.Msg {
ctx := context.Background()
certPEM, err := client.GetUserSSLCert(ctx, uuid, false, false)
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

// View renders the cert detail.
func (c *CertDetail) View() string {
if c.Cert == nil {
return tui.MutedStyle.Render("No certificate selected.")
}

switch c.mode {
case certDetailAnalysis:
return c.viewAnalysis()
case certDetailViewCert:
return c.viewCertContent()
case certDetailChainSel:
return c.viewChainSelector()
case certDetailExportPath:
return c.viewExportPath()
case certDetailPrivKeyPass:
return c.viewPrivKeyPass()
case certDetailViewPrivKey:
return c.viewPrivKey()
case certDetailExportPriv:
return c.viewExportPriv()
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

if c.privMsg != "" {
sb.WriteString("\n")
sb.WriteString(tui.DangerStyle.Render(c.privMsg))
sb.WriteString("\n")
}

sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("a: analyze • v: view cert • e: export cert • k: view privkey • K: export privkey • esc: back"))
return sb.String()
}

func (c *CertDetail) viewAnalysis() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("📜 SSL Certificate Analysis"))
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

func (c *CertDetail) viewCertContent() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("📜 SSL Certificate Content"))
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

func (c *CertDetail) viewChainSelector() string {
var sb strings.Builder
actionLabel := "View"
if c.chainAction == certDetailActionExport {
actionLabel = "Export"
}
sb.WriteString(tui.TitleStyle.Render(fmt.Sprintf("📜 %s Certificate — Select Type", actionLabel)))
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

func (c *CertDetail) viewExportPath() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("📜 Export Certificate"))
sb.WriteString("\n\n")
sb.WriteString(tui.NormalStyle.Render("Enter the file path to save the certificate:"))
sb.WriteString("\n")
sb.WriteString(tui.InputFocusStyle.Width(c.width - 4).Render(c.exportInput.InputView()))
sb.WriteString("\n")
if s := c.exportInput.SuggestionsView(); s != "" {
sb.WriteString(s)
}
sb.WriteString("\n")
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
sb.WriteString(tui.HelpStyle.Render("tab: autocomplete • enter: save • esc: back"))
return sb.String()
}

func (c *CertDetail) viewPrivKeyPass() string {
var sb strings.Builder
action := "View"
if c.privKeyMode == certDetailActionExport {
action = "Export"
}
sb.WriteString(tui.TitleStyle.Render(fmt.Sprintf("🔑 %s Private Key — Authentication", action)))
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

func (c *CertDetail) viewPrivKey() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("🔑 SSL Certificate Private Key"))
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

func (c *CertDetail) viewExportPriv() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("🔑 Export Private Key"))
sb.WriteString("\n\n")
sb.WriteString(tui.NormalStyle.Render("Enter the file path to save the private key:"))
sb.WriteString("\n")
sb.WriteString(tui.InputFocusStyle.Width(c.width - 4).Render(c.privExport.InputView()))
sb.WriteString("\n")
if s := c.privExport.SuggestionsView(); s != "" {
sb.WriteString(s)
}
sb.WriteString("\n")
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
sb.WriteString(tui.HelpStyle.Render("tab: autocomplete • enter: save • esc: back"))
return sb.String()
}
