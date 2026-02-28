package views

import (
"context"
"strings"

"github.com/charmbracelet/bubbles/textinput"
tea "github.com/charmbracelet/bubbletea"
"github.com/charmbracelet/lipgloss"
"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
"github.com/gregPerlinLi/CertVaultCLIX/internal/config"
tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/components"
)

const banner = `
 ██████╗███████╗██████╗ ████████╗██╗   ██╗ █████╗ ██╗   ██╗██╗  ████████╗
██╔════╝██╔════╝██╔══██╗╚══██╔══╝██║   ██║██╔══██╗██║   ██║██║  ╚══██╔══╝
██║     █████╗  ██████╔╝   ██║   ██║   ██║███████║██║   ██║██║     ██║
██║     ██╔══╝  ██╔══██╗   ██║   ╚██╗ ██╔╝██╔══██║██║   ██║██║     ██║
╚██████╗███████╗██║  ██║   ██║    ╚████╔╝ ██║  ██║╚██████╔╝███████╗██║
 ╚═════╝╚══════╝╚═╝  ╚═╝   ╚═╝     ╚═══╝  ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  X
`

// connRefusedHint is shown when a login attempt fails with a network-connectivity
// error so users are guided to install and configure the CertVault server.
const connRefusedHint = "Please install and start the CertVault server first, then\n" +
	"set the correct server URL (ctrl+u).\n" +
	"→ https://github.com/gregPerlinLi/CertVault"

// LoginSuccessMsg is sent after a successful login.
type LoginSuccessMsg struct {
Profile *api.UserProfile
}

// LoginErrorMsg is sent after a failed login.
type LoginErrorMsg struct {
Err error
}

// Login is the login screen view.
type Login struct {
client     *api.Client
cfg        *config.Config
usernameIn textinput.Model
passwordIn textinput.Model
serverIn   textinput.Model
// focused: 0=username, 1=password, 2=button, 3=serverURL(edit mode)
focused     int
editingURL  bool
loading     bool
spinner     components.Spinner
err         string
hint        string // extra hint shown below the error (e.g. connection-refused guidance)
width       int
height      int
}

// NewLogin creates a new Login view.
func NewLogin(client *api.Client, cfg *config.Config) Login {
u := textinput.New()
u.Placeholder = "Username"
u.Focus()
u.CharLimit = 64

p := textinput.New()
p.Placeholder = "Password"
p.EchoMode = textinput.EchoPassword
p.CharLimit = 128

s := textinput.New()
s.Placeholder = config.DefaultServerURL
s.CharLimit = 256

serverURL := config.DefaultServerURL
if cfg != nil && cfg.ServerURL != "" {
serverURL = cfg.ServerURL
}
s.SetValue(serverURL)

return Login{
client:     client,
cfg:        cfg,
usernameIn: u,
passwordIn: p,
serverIn:   s,
spinner:    components.NewSpinner(),
}
}

// SetSize updates the view dimensions.
func (l *Login) SetSize(width, height int) {
l.width = width
l.height = height
}

// Init initializes the login view.
func (l *Login) Init() tea.Cmd {
return textinput.Blink
}

// Update handles messages.
func (l *Login) Update(msg tea.Msg) (tea.Cmd, bool) {
switch msg := msg.(type) {
case tea.KeyMsg:
if l.loading {
return l.spinner.Update(msg), false
}
if l.editingURL {
switch msg.String() {
case "enter":
newURL := l.serverIn.Value()
if newURL != "" {
l.client.SetBaseURL(newURL)
if l.cfg != nil {
l.cfg.ServerURL = newURL
_ = config.Save(l.cfg)
}
}
l.editingURL = false
l.serverIn.Blur()
l.usernameIn.Focus()
l.focused = 0
return nil, false
case "esc":
l.editingURL = false
l.serverIn.Blur()
// Restore original server URL from config (only if config is available)
if l.cfg != nil {
l.serverIn.SetValue(l.cfg.ServerURL)
}
l.usernameIn.Focus()
l.focused = 0
return nil, false
}
var cmd tea.Cmd
l.serverIn, cmd = l.serverIn.Update(msg)
return cmd, false
}
switch msg.String() {
case "tab", "down":
l.nextField()
return nil, false
case "shift+tab", "up":
l.prevField()
return nil, false
case "enter":
if l.focused == 2 {
return l.doLogin(), false
}
l.nextField()
return nil, false
case "ctrl+u":
// Open server URL edit (Ctrl+U, safe to use anywhere).
l.editingURL = true
l.usernameIn.Blur()
l.passwordIn.Blur()
l.serverIn.Focus()
return nil, false
}

case LoginSuccessMsg:
return nil, true

case LoginErrorMsg:
l.loading = false
l.spinner.Stop()
if msg.Err != nil {
l.err = msg.Err.Error()
// Detect connection-refused errors and guide the user to the server.
if strings.Contains(strings.ToLower(l.err), "connection refused") ||
strings.Contains(strings.ToLower(l.err), "no such host") ||
strings.Contains(strings.ToLower(l.err), "dial tcp") {
l.hint = connRefusedHint
} else {
l.hint = ""
}
}
return nil, false
}

var cmd tea.Cmd
switch l.focused {
case 0:
l.usernameIn, cmd = l.usernameIn.Update(msg)
case 1:
l.passwordIn, cmd = l.passwordIn.Update(msg)
default:
cmd = l.spinner.Update(msg)
}
return cmd, false
}

func (l *Login) nextField() {
switch l.focused {
case 0:
l.usernameIn.Blur()
l.passwordIn.Focus()
l.focused = 1
case 1:
l.passwordIn.Blur()
l.focused = 2
case 2:
l.usernameIn.Focus()
l.focused = 0
}
}

func (l *Login) prevField() {
switch l.focused {
case 2:
l.passwordIn.Focus()
l.focused = 1
case 1:
l.passwordIn.Blur()
l.usernameIn.Focus()
l.focused = 0
case 0:
l.usernameIn.Blur()
l.focused = 2
}
}

func (l *Login) doLogin() tea.Cmd {
l.loading = true
l.err = ""
l.hint = ""
username := l.usernameIn.Value()
password := l.passwordIn.Value()

spinCmd := l.spinner.Start("Logging in...")

return tea.Batch(spinCmd, func() tea.Msg {
if err := l.client.Login(context.Background(), username, password); err != nil {
return LoginErrorMsg{Err: err}
}
profile, err := l.client.GetProfile(context.Background())
if err != nil {
return LoginErrorMsg{Err: err}
}
// Save session to config
if l.cfg != nil {
l.cfg.Session = l.client.GetSession()
_ = config.Save(l.cfg)
}
return LoginSuccessMsg{Profile: profile}
})
}

// View renders the login screen.
func (l *Login) View() string {
var sb strings.Builder

// Banner
bannerLines := strings.Split(strings.TrimSpace(banner), "\n")
for _, line := range bannerLines {
sb.WriteString(tui.TitleStyle.Render(line))
sb.WriteString("\n")
}
sb.WriteString("\n")

// Subtitle
sb.WriteString(tui.SubtitleStyle.Render("CertVault CLI Extended — Interactive TUI"))
sb.WriteString("\n\n")

formWidth := 44

// Server URL display / edit
serverURL := l.serverIn.Value()
if l.editingURL {
sb.WriteString(tui.NormalStyle.Render("Server URL:"))
sb.WriteString("\n")
sb.WriteString(tui.InputFocusStyle.Width(formWidth).Render(l.serverIn.View()))
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("  enter: confirm • esc: cancel"))
sb.WriteString("\n\n")
} else {
sb.WriteString(tui.MutedStyle.Render("Server: " + serverURL))
sb.WriteString(tui.HelpStyle.Render("  [ctrl+u] change"))
sb.WriteString("\n\n")
}

// Username
var userStyle lipgloss.Style
if l.focused == 0 {
userStyle = tui.InputFocusStyle
} else {
userStyle = tui.InputStyle
}
sb.WriteString(tui.NormalStyle.Render("Username:"))
sb.WriteString("\n")
sb.WriteString(userStyle.Width(formWidth).Render(l.usernameIn.View()))
sb.WriteString("\n\n")

// Password
var passStyle lipgloss.Style
if l.focused == 1 {
passStyle = tui.InputFocusStyle
} else {
passStyle = tui.InputStyle
}
sb.WriteString(tui.NormalStyle.Render("Password:"))
sb.WriteString("\n")
sb.WriteString(passStyle.Width(formWidth).Render(l.passwordIn.View()))
sb.WriteString("\n\n")

// Login button
var btnStyle lipgloss.Style
if l.focused == 2 {
btnStyle = tui.ButtonStyle
} else {
btnStyle = tui.ButtonInactiveStyle
}
sb.WriteString(btnStyle.Render("  Login  "))
sb.WriteString("\n\n")

// Loading / spinner
if l.loading {
sb.WriteString(l.spinner.View())
sb.WriteString("\n")
}

// Error
if l.err != "" {
// Wrap to formWidth so long errors don't widen the centering computation.
wrapped := wrapText(l.err, formWidth-2, "  ")
for i, line := range strings.Split(wrapped, "\n") {
if i == 0 {
sb.WriteString(tui.DangerStyle.Render("✗ " + line))
} else {
sb.WriteString(tui.DangerStyle.Render(line))
}
sb.WriteString("\n")
}
}

// Hint (e.g. connection-refused guidance)
if l.hint != "" {
for _, line := range strings.Split(l.hint, "\n") {
sb.WriteString(tui.HelpStyle.Render("  " + line))
sb.WriteString("\n")
}
}

// Help
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("tab/↓ next • shift+tab/↑ prev • enter select • ctrl+u server URL • ctrl+q quit"))

content := sb.String()
// Center the content.
// We compute the widest "stable" line (banner / form elements) and use that for
// horizontal centering. Lines introduced by dynamic content (error, hint, spinner)
// are capped at formWidth so they never push the layout to the left.
lines := strings.Split(content, "\n")
maxW := 0
for _, line := range lines {
w := lipgloss.Width(line)
if w > maxW {
maxW = w
}
}
// Cap at a sane maximum (banner + padding is ~80 cols) so dynamic error text
// that was wrapped to formWidth doesn't distort the centering.
const maxContentWidth = 82
if maxW > maxContentWidth {
maxW = maxContentWidth
}

padLeft := (l.width - maxW) / 2
if padLeft < 0 {
padLeft = 0
}
padTop := (l.height - len(lines)) / 2
if padTop < 0 {
padTop = 0
}

leftPad := strings.Repeat(" ", padLeft)
topPad := strings.Repeat("\n", padTop)

var out strings.Builder
out.WriteString(topPad)
for _, line := range lines {
out.WriteString(leftPad)
out.WriteString(line)
out.WriteString("\n")
}
return out.String()
}
