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
RoleInt  int // numeric role for color coding
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
// Render the role with its color on the status bar background
roleColored := st.RoleStyle(s.RoleInt).
Background(st.ColorPrimary).
Bold(true).
Render("(" + s.Role + ")")

// Build left section: icon + username + colored role
leftPlain := fmt.Sprintf(" 🔐 %s ", s.Username)
left := st.StatusBarHighlight.Render(leftPlain) + roleColored + st.StatusBarHighlight.Render(" ")

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
