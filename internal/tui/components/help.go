package components

import (
	"strings"

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
		{Key: "↑/k, ↓/j", Desc: "Move up/down"},
		{Key: "←/h, →/l", Desc: "Navigate panels"},
		{Key: "enter", Desc: "Select / confirm"},
		{Key: "esc", Desc: "Back / cancel"},
		{Key: "r/F5", Desc: "Refresh"},
		{Key: "/", Desc: "Search"},
		{Key: "n", Desc: "New item"},
		{Key: "d", Desc: "Delete item"},
		{Key: "e", Desc: "Edit item"},
		{Key: "x", Desc: "Export"},
		{Key: "tab", Desc: "Next field"},
		{Key: "?", Desc: "Toggle help"},
		{Key: "q", Desc: "Quit"},
	}
}

// View renders the help overlay.
func (h *Help) View() string {
	if !h.visible {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(st.TitleStyle.Render("Keyboard Shortcuts"))
	sb.WriteString("\n\n")
	for _, e := range h.Entries {
		key := st.KeyStyle.Render(e.Key)
		desc := st.HelpStyle.Render(e.Desc)
		sb.WriteString("  " + key + "  " + desc + "\n")
	}
	sb.WriteString("\n")
	sb.WriteString(st.MutedStyle.Render("Press ? to close"))

	w := h.width / 2
	if w < 50 {
		w = 50
	}
	return st.DialogStyle.Width(w).Render(sb.String())
}
