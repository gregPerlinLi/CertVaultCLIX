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
certDetailNormal    certDetailMode = iota
certDetailAnalysis                 // showing analysis result
certDetailViewCert                 // viewing cert PEM content
certDetailChainSel                 // selecting chain type before view/export
certDetailExportPath               // entering export path
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
certContent  string // decoded PEM content
chainSelIdx  int
chainAction  certDetailAction
// export path input
exportInput  textinput.Model
exportMsg    string
width        int
height       int
}

// NewCertDetail creates a new SSL cert detail view.
func NewCertDetail(cert *api.SSLCert, client *api.Client) CertDetail {
vp := viewport.New(80, 20)
ei := textinput.New()
ei.Placeholder = "e.g. /home/user/cert.pem"
ei.CharLimit = 512
return CertDetail{
Cert:        cert,
client:      client,
spinner:     components.NewSpinner(),
resultVP:    vp,
exportInput: ei,
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

// IsAnalysisMode returns true when the inline analysis result is being shown.
func (c *CertDetail) IsAnalysisMode() bool {
return c.mode == certDetailAnalysis ||
c.mode == certDetailViewCert ||
c.mode == certDetailChainSel ||
c.mode == certDetailExportPath
}

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
// Switch to export path input.
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
// export: ask for path
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
var cmd tea.Cmd
c.exportInput, cmd = c.exportInput.Update(msg)
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
}
}
case tea.MouseMsg:
if c.mode == certDetailAnalysis || c.mode == certDetailViewCert {
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
}
return c.spinner.Update(msg)
}

// certContentMsg carries decoded PEM content.
type certContentMsg struct {
content string
err     string
}

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

func (c *CertDetail) doExport() tea.Cmd {
path := strings.TrimSpace(c.exportInput.Value())
if path == "" {
c.exportMsg = "Path cannot be empty"
return nil
}
content := c.certContent
if content == "" {
// Need to fetch first — get chain option then export.
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
if err := writeFile(path, []byte(content)); err != nil {
c.exportMsg = "Export failed: " + err.Error()
return nil
}
c.exportMsg = "✓ Saved to " + path
c.exportInput.Blur()
c.mode = certDetailViewCert
return nil
}

// certExportedMsg signals a successful export when content was fetched during export.
type certExportedMsg struct{ path string }

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

sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("a: analyze • v: view cert • e: export cert • esc: back"))
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
