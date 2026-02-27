package components

import (
	"strings"

	st "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
)

// SidebarItem is a single item in the navigation sidebar.
type SidebarItem struct {
	Icon     string
	Label    string
	ID       string
	Children []SidebarItem
}

// Sidebar is the navigation sidebar component.
type Sidebar struct {
	Items    []SidebarItem
	cursor   int
	focused  bool
	width    int
	height   int
}

// NewSidebar creates a new sidebar.
func NewSidebar(items []SidebarItem) Sidebar {
	return Sidebar{
		Items:   items,
		focused: true,
	}
}

// SetSize sets the sidebar dimensions.
func (s *Sidebar) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// SetFocused sets the focus state.
func (s *Sidebar) SetFocused(f bool) {
	s.focused = f
}

// SelectedID returns the ID of the currently selected item.
func (s *Sidebar) SelectedID() string {
	if len(s.Items) == 0 || s.cursor < 0 || s.cursor >= len(s.Items) {
		return ""
	}
	return s.Items[s.cursor].ID
}

// SelectedIndex returns the cursor index.
func (s *Sidebar) SelectedIndex() int {
	return s.cursor
}

// SetCursor sets the cursor to the given index.
func (s *Sidebar) SetCursor(i int) {
	if i >= 0 && i < len(s.Items) {
		s.cursor = i
	}
}

// MoveUp moves the cursor up.
func (s *Sidebar) MoveUp() {
	if s.cursor > 0 {
		s.cursor--
	}
}

// MoveDown moves the cursor down.
func (s *Sidebar) MoveDown() {
	if s.cursor < len(s.Items)-1 {
		s.cursor++
	}
}

// View renders the sidebar.
func (s *Sidebar) View() string {
	var sb strings.Builder
	sb.WriteString(st.SidebarHeaderStyle.Render("Navigation"))
	sb.WriteString("\n\n")

	for i, item := range s.Items {
		icon := item.Icon
		if icon == "" {
			icon = "•"
		}
		label := icon + " " + item.Label

		var line string
		if s.focused && i == s.cursor {
			line = st.SidebarSelectedStyle.Render(label)
		} else {
			line = st.SidebarItemStyle.Render(label)
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return st.SidebarStyle.Width(s.width).Height(s.height).Render(sb.String())
}
