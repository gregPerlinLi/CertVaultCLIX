package views

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
	tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/components"
)

// SuperadminMode indicates which superadmin sub-view is active.
type SuperadminMode int

const (
	SuperadminModeMenu SuperadminMode = iota
	SuperadminModeSessions
	SuperadminModeUsers
)

// SuperadminDataMsg carries superadmin data.
type SuperadminDataMsg struct {
	Sessions []api.AllSession
	Users    []api.AdminUser
	Total    int64
	Err      error
}

// Superadmin is the superadmin management view.
type Superadmin struct {
	client  *api.Client
	mode    SuperadminMode
	menuIdx int
	table   components.Table
	sessions []api.AllSession
	users   []api.AdminUser
	total   int64
	page    int
	spinner components.Spinner
	toast   components.Toast
	dialog  *components.Dialog
	err     string
	width   int
	height  int
}

var superadminMenuItems = []string{
	"All Sessions",
	"User Management",
}

// NewSuperadmin creates a new superadmin view.
func NewSuperadmin(client *api.Client) Superadmin {
	cols := []components.Column{
		{Title: "Username", Width: 20},
		{Title: "IP", Width: 18},
		{Title: "User Agent", Width: 30},
		{Title: "Login At", Width: 20},
	}
	return Superadmin{
		client:  client,
		table:   components.NewTable(cols, 15),
		page:    1,
		spinner: components.NewSpinner(),
	}
}

// SetSize updates dimensions.
func (s *Superadmin) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.table.SetSize(width, height-5)
}

// Init initializes.
func (s *Superadmin) Init() tea.Cmd { return nil }

// Update handles messages.
func (s *Superadmin) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case SuperadminDataMsg:
		s.spinner.Stop()
		if msg.Err != nil {
			s.err = msg.Err.Error()
			return nil
		}
		s.sessions = msg.Sessions
		s.users = msg.Users
		s.total = msg.Total
		if s.mode == SuperadminModeSessions {
			s.table.SetRows(s.buildSessionRows())
		} else {
			s.table.SetRows(s.buildUserRows())
		}
		return nil

	case components.ConfirmMsg:
		s.dialog = nil
		if msg.Confirmed {
			return s.forceLogout()
		}
		return nil

	case components.ClearToastMsg:
		s.toast.Hide()
		return nil

	case tea.KeyMsg:
		if s.dialog != nil {
			cmd, _ := s.dialog.Update(msg)
			return cmd
		}
		switch s.mode {
		case SuperadminModeMenu:
			switch msg.String() {
			case "up", "k":
				if s.menuIdx > 0 {
					s.menuIdx--
				}
			case "down", "j":
				if s.menuIdx < len(superadminMenuItems)-1 {
					s.menuIdx++
				}
			case "enter":
				s.mode = SuperadminMode(s.menuIdx + 1)
				s.page = 1
				// Set appropriate columns
				if s.mode == SuperadminModeUsers {
					s.table.Columns = []components.Column{
						{Title: "Username", Width: 20},
						{Title: "Display Name", Width: 25},
						{Title: "Email", Width: 30},
						{Title: "Role", Width: 12},
					}
				} else {
					s.table.Columns = []components.Column{
						{Title: "Username", Width: 20},
						{Title: "IP", Width: 18},
						{Title: "User Agent", Width: 30},
						{Title: "Login At", Width: 20},
					}
				}
				cmd := s.spinner.Start("Loading...")
				return tea.Batch(cmd, s.load())
			}
		default:
			switch msg.String() {
			case "esc":
				s.mode = SuperadminModeMenu
			case "r", "f5":
				cmd := s.spinner.Start("Refreshing...")
				return tea.Batch(cmd, s.load())
			case "d", "delete":
				if s.mode == SuperadminModeSessions {
					idx := s.table.SelectedIndex()
					if idx >= 0 && idx < len(s.sessions) {
						d := components.NewDialog("Force Logout",
							fmt.Sprintf("Force logout user %s?", s.sessions[idx].Username))
						s.dialog = &d
					}
				}
			default:
				return s.table.Update(msg)
			}
		}
	}
	return s.spinner.Update(msg)
}

func (s *Superadmin) load() tea.Cmd {
	mode := s.mode
	page := s.page
	return func() tea.Msg {
		ctx := context.Background()
		switch mode {
		case SuperadminModeSessions:
			sessions, err := s.client.ListAllSessions(ctx, page, 20)
			if err != nil {
				return SuperadminDataMsg{Err: err}
			}
			return SuperadminDataMsg{Sessions: sessions.List, Total: sessions.Total}
		case SuperadminModeUsers:
			users, err := s.client.ListAdminUsers(ctx, page, 20)
			if err != nil {
				return SuperadminDataMsg{Err: err}
			}
			return SuperadminDataMsg{Users: users.List, Total: users.Total}
		}
		return SuperadminDataMsg{}
	}
}

func (s *Superadmin) forceLogout() tea.Cmd {
	idx := s.table.SelectedIndex()
	if idx < 0 || idx >= len(s.sessions) {
		return nil
	}
	username := s.sessions[idx].Username
	return func() tea.Msg {
		err := s.client.ForceLogoutUser(context.Background(), username)
		if err != nil {
			return SuperadminDataMsg{Err: err}
		}
		sessions, err := s.client.ListAllSessions(context.Background(), s.page, 20)
		if err != nil {
			return SuperadminDataMsg{Err: err}
		}
		return SuperadminDataMsg{Sessions: sessions.List, Total: sessions.Total}
	}
}

func (s *Superadmin) buildSessionRows() []components.Row {
	rows := make([]components.Row, len(s.sessions))
	for i, sess := range s.sessions {
		rows[i] = components.Row{
			sess.Username,
			sess.IP,
			truncate(sess.UserAgent, 30),
			sess.LoginAt,
		}
	}
	return rows
}

func (s *Superadmin) buildUserRows() []components.Row {
	rows := make([]components.Row, len(s.users))
	for i, u := range s.users {
		rows[i] = components.Row{
			u.Username,
			u.DisplayName,
			u.Email,
			api.RoleName(u.Role),
		}
	}
	return rows
}

// View renders the superadmin view.
func (s *Superadmin) View() string {
	var sb strings.Builder
	sb.WriteString(tui.TitleStyle.Render("👑 Superadmin"))
	sb.WriteString("\n\n")

	if s.dialog != nil {
		return sb.String() + s.dialog.View(s.width)
	}

	if s.mode == SuperadminModeMenu {
		for i, item := range superadminMenuItems {
			if i == s.menuIdx {
				sb.WriteString(tui.SelectedStyle.Render("▶ " + item))
			} else {
				sb.WriteString(tui.NormalStyle.Render("  " + item))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
		sb.WriteString(tui.HelpStyle.Render("↑/↓: select • enter: open"))
		return sb.String()
	}

	title := superadminMenuItems[s.mode-1]
	sb.WriteString(tui.SubtitleStyle.Render(title))
	sb.WriteString("\n\n")

	if s.spinner.IsActive() {
		sb.WriteString(s.spinner.View())
		return sb.String()
	}
	if s.err != "" {
		sb.WriteString(tui.DangerStyle.Render("Error: " + s.err))
		sb.WriteString("\n")
		sb.WriteString(tui.HelpStyle.Render("Press r to retry • esc: back"))
		return sb.String()
	}

	total := fmt.Sprintf("Total: %d | Page: %d", s.total, s.page)
	sb.WriteString(tui.MutedStyle.Render(total))
	sb.WriteString("\n")
	sb.WriteString(s.table.View())
	sb.WriteString("\n")
	sb.WriteString(tui.HelpStyle.Render("d: force logout • r: refresh • esc: back • PgUp/PgDn: page"))

	if s.toast.IsVisible() {
		sb.WriteString("\n" + s.toast.View())
	}
	return sb.String()
}
