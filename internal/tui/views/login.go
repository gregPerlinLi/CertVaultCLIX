package views

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
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
	client      *api.Client
	usernameIn  textinput.Model
	passwordIn  textinput.Model
	focused     int // 0=username, 1=password, 2=button
	loading     bool
	spinner     components.Spinner
	err         string
	width       int
	height      int
}

// NewLogin creates a new Login view.
func NewLogin(client *api.Client) Login {
	u := textinput.New()
	u.Placeholder = "Username"
	u.Focus()
	u.CharLimit = 64

	p := textinput.New()
	p.Placeholder = "Password"
	p.EchoMode = textinput.EchoPassword
	p.CharLimit = 128

	return Login{
		client:     client,
		usernameIn: u,
		passwordIn: p,
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
		}

	case LoginSuccessMsg:
		return nil, true

	case LoginErrorMsg:
		l.loading = false
		l.spinner.Stop()
		if msg.Err != nil {
			l.err = msg.Err.Error()
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

	// Form
	var userStyle, passStyle lipgloss.Style
	if l.focused == 0 {
		userStyle = tui.InputFocusStyle
	} else {
		userStyle = tui.InputStyle
	}
	if l.focused == 1 {
		passStyle = tui.InputFocusStyle
	} else {
		passStyle = tui.InputStyle
	}

	formWidth := 40
	sb.WriteString(tui.NormalStyle.Render("Username:"))
	sb.WriteString("\n")
	sb.WriteString(userStyle.Width(formWidth).Render(l.usernameIn.View()))
	sb.WriteString("\n\n")

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

	// Loading
	if l.loading {
		sb.WriteString(l.spinner.View())
		sb.WriteString("\n")
	}

	// Error
	if l.err != "" {
		sb.WriteString(tui.DangerStyle.Render("✗ " + l.err))
		sb.WriteString("\n")
	}

	// Help
	sb.WriteString("\n")
	sb.WriteString(tui.HelpStyle.Render("tab/↓ next field • shift+tab/↑ prev field • enter select • q quit"))

	content := sb.String()
	// Center the content
	lines := strings.Split(content, "\n")
	maxW := 0
	for _, line := range lines {
		w := lipgloss.Width(line)
		if w > maxW {
			maxW = w
		}
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
