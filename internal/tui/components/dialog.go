package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	st "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
)

// ConfirmMsg is sent when the user confirms or cancels a dialog.
type ConfirmMsg struct {
	Confirmed bool
}

// Dialog is a modal confirmation dialog.
type Dialog struct {
	Title   string
	Message string
	confirm int // 0 = yes, 1 = no
}

// NewDialog creates a new dialog.
func NewDialog(title, message string) Dialog {
	return Dialog{
		Title:   title,
		Message: message,
		confirm: 1, // default to "no"
	}
}

// Update handles keyboard events for the dialog.
func (d *Dialog) Update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "tab":
			d.confirm = 0
		case "right", "l":
			d.confirm = 1
		case "enter":
			return func() tea.Msg { return ConfirmMsg{Confirmed: d.confirm == 0} }, true
		case "esc", "q":
			return func() tea.Msg { return ConfirmMsg{Confirmed: false} }, true
		}
	}
	return nil, false
}

// WasConfirmed returns whether the current dialog selection is "Yes".
// Call this after Update returns done=true to read the user's choice without
// relying on the ConfirmMsg being dispatched through the bubbletea message loop.
func (d *Dialog) WasConfirmed() bool {
	return d.confirm == 0
}
func (d *Dialog) View(width int) string {
	var sb strings.Builder

	sb.WriteString(st.TitleStyle.Render(d.Title))
	sb.WriteString("\n\n")
	sb.WriteString(st.NormalStyle.Render(d.Message))
	sb.WriteString("\n\n")

	var yes, no lipgloss.Style
	if d.confirm == 0 {
		yes = st.ButtonStyle
		no = st.ButtonInactiveStyle
	} else {
		yes = st.ButtonInactiveStyle
		no = st.ButtonStyle
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center,
		yes.Render(" Yes "),
		"   ",
		no.Render(" No "),
	)
	sb.WriteString(buttons)

	inner := sb.String()
	dialogWidth := width / 2
	if dialogWidth < 40 {
		dialogWidth = 40
	}

	return st.DialogStyle.Width(dialogWidth).Render(inner)
}
