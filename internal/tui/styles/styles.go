package styles

import "github.com/charmbracelet/lipgloss"

// Theme colors
const (
	ColorPrimary   = lipgloss.Color("#7C3AED")
	ColorSecondary = lipgloss.Color("#A78BFA")
	ColorSuccess   = lipgloss.Color("#10B981")
	ColorWarning   = lipgloss.Color("#F59E0B")
	ColorDanger    = lipgloss.Color("#EF4444")
	ColorMuted     = lipgloss.Color("#6B7280")
	ColorBg        = lipgloss.Color("#1E1B2E")
	ColorSurface   = lipgloss.Color("#2D2B3D")
	ColorBorder    = lipgloss.Color("#4C4A6B")
	ColorText      = lipgloss.Color("#E2E8F0")
	ColorTextMuted = lipgloss.Color("#94A3B8")
	ColorHighlight = lipgloss.Color("#EDE9FE")
	ColorWhite     = lipgloss.Color("#FFFFFF")
)

// Styles are the application-wide lipgloss styles.
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Padding(0, 1)

	SelectedStyle = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(ColorWhite).
			Bold(true).
			Padding(0, 1)

	NormalStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Padding(0, 1)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Padding(0, 1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	DangerStyle = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	FocusBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1)

	StatusBarStyle = lipgloss.NewStyle().
			Background(ColorSurface).
			Foreground(ColorText).
			Padding(0, 1)

	StatusBarHighlight = lipgloss.NewStyle().
				Background(ColorPrimary).
				Foreground(ColorWhite).
				Bold(true).
				Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	KeyStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	PaginationStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	SidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	SidebarItemStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Padding(0, 1)

	SidebarSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Padding(0, 1)

	SidebarHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				Bold(true).
				Padding(0, 1)

	DialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Background(ColorSurface).
			Padding(1, 2)

	ButtonStyle = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(ColorWhite).
			Bold(true).
			Padding(0, 2)

	ButtonInactiveStyle = lipgloss.NewStyle().
				Background(ColorSurface).
				Foreground(ColorTextMuted).
				Padding(0, 2)

	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	InputFocusStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorSecondary)

	TableRowStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	TableSelectedRowStyle = lipgloss.NewStyle().
				Background(ColorPrimary).
				Foreground(ColorWhite)

	ToastSuccessStyle = lipgloss.NewStyle().
				Background(ColorSuccess).
				Foreground(ColorWhite).
				Bold(true).
				Padding(0, 2)

	ToastErrorStyle = lipgloss.NewStyle().
			Background(ColorDanger).
			Foreground(ColorWhite).
			Bold(true).
			Padding(0, 2)

	ToastInfoStyle = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(ColorWhite).
			Bold(true).
			Padding(0, 2)
)

// ExpiryStyle returns a color-coded style for a certificate expiry date.
// Green (>30 days), Yellow (<30 days), Red (expired).
func ExpiryStyle(daysLeft int) lipgloss.Style {
	switch {
	case daysLeft < 0:
		return DangerStyle
	case daysLeft < 30:
		return WarningStyle
	default:
		return SuccessStyle
	}
}

// RoleStyle returns a color-coded style for a user role.
// User = green, Admin = blue/purple (secondary), Superadmin = red.
func RoleStyle(role int) lipgloss.Style {
	switch role {
	case 1: // User
		return lipgloss.NewStyle().Foreground(ColorSuccess)
	case 2: // Admin
		return lipgloss.NewStyle().Foreground(ColorSecondary)
	case 3: // Superadmin
		return lipgloss.NewStyle().Foreground(ColorDanger)
	default:
		return lipgloss.NewStyle().Foreground(ColorTextMuted)
	}
}
