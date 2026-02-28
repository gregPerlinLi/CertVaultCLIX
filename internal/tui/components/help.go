package components

import (
"fmt"
"strings"

"github.com/charmbracelet/lipgloss"
st "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
)

// HelpEntry is a single help line.
type HelpEntry struct {
Key  string
Desc string
}

// Help renders a keyboard shortcut overlay.
type Help struct {
Entries []HelpEntry
visible bool
width   int
}

// NewHelp creates a new help overlay.
func NewHelp(entries []HelpEntry) Help {
return Help{Entries: entries}
}

// SetSize sets the help overlay width.
func (h *Help) SetSize(width int) {
h.width = width
}

// Toggle toggles visibility.
func (h *Help) Toggle() {
h.visible = !h.visible
}

// IsVisible returns visibility.
func (h *Help) IsVisible() bool {
return h.visible
}

// DefaultEntries returns the default global help entries.
func DefaultEntries() []HelpEntry {
return []HelpEntry{
{Key: "↑/k, ↓/j", Desc: "Move up/down in list"},
{Key: "ctrl+u/d", Desc: "Scroll table by half-page"},
{Key: "ctrl+l", Desc: "Clear input field"},
{Key: "scroll/drag", Desc: "Mouse wheel navigation"},
{Key: "[/]", Desc: "Prev/next API page"},
{Key: "enter", Desc: "Select / confirm"},
{Key: "esc", Desc: "Back / cancel"},
{Key: "r/F5", Desc: "Refresh"},
{Key: "n", Desc: "New item"},
{Key: "d", Desc: "Delete item"},
{Key: "x", Desc: "Export"},
{Key: "tab", Desc: "Next field"},
{Key: "?", Desc: "Toggle help"},
{Key: "q", Desc: "Quit (from list/detail views)"},
{Key: "ctrl+c", Desc: "Force quit"},
{Key: "L (sidebar)", Desc: "Logout with confirmation"},
}
}

// View renders the help overlay.
func (h *Help) View() string {
if !h.visible {
return ""
}

w := h.width / 2
if w < 54 {
w = 54
}
// Content width = dialog width - border(2) - padding(4)
contentW := w - 6
if contentW < 20 {
contentW = 20
}

var sb strings.Builder
sb.WriteString(st.TitleStyle.Render("Keyboard Shortcuts"))
sb.WriteString("\n\n")

keyW := 14
for _, e := range h.Entries {
keyPart := fmt.Sprintf("%-*s", keyW, e.Key)
line := lipgloss.NewStyle().
Foreground(st.ColorSecondary).
Bold(true).
Render(keyPart) +
"  " +
lipgloss.NewStyle().
Foreground(st.ColorText).
Background(st.ColorSurface).
Width(contentW - keyW - 2).
Render(e.Desc)
sb.WriteString(line)
sb.WriteString("\n")
}
sb.WriteString("\n")
sb.WriteString(st.MutedStyle.Render("Press ? to close"))

return st.DialogStyle.Width(w).Render(sb.String())
}
