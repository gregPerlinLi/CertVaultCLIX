package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	st "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
)

// StatusBar renders a bottom status bar.
type StatusBar struct {
	Username string
	Role     string
	Server   string
	Status   string
	width    int
}

// SetSize sets the status bar width.
func (s *StatusBar) SetSize(width int) {
	s.width = width
}

// View renders the status bar.
func (s *StatusBar) View() string {
	left := st.StatusBarHighlight.Render(fmt.Sprintf(" 🔐 %s (%s) ", s.Username, s.Role))
	right := st.StatusBarStyle.Render(fmt.Sprintf(" %s ", s.Server))

	mid := ""
	if s.Status != "" {
		mid = st.StatusBarStyle.Render(fmt.Sprintf("  %s  ", s.Status))
	}

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	midWidth := lipgloss.Width(mid)

	remaining := s.width - leftWidth - rightWidth - midWidth
	if remaining < 0 {
		remaining = 0
	}
	padding := strings.Repeat(" ", remaining)

	return left + mid + st.StatusBarStyle.Render(padding) + right
}
