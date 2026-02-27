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

// AdminMode indicates which admin sub-view is active.
type AdminMode int

const (
	AdminModeMenu AdminMode = iota
	AdminModeUsers
	AdminModeCAs
)

// AdminDataMsg carries admin data.
type AdminDataMsg struct {
	Users []api.AdminUser
	CAs   []api.CACert
	Total int64
	Err   error
}

// Admin is the admin management view.
type Admin struct {
	client  *api.Client
	mode    AdminMode
	menuIdx int
	table   components.Table
	users   []api.AdminUser
	cas     []api.CACert
	total   int64
	page    int
	spinner components.Spinner
	toast   components.Toast
	err     string
	width   int
	height  int
}

var adminMenuItems = []string{
	"User Management",
	"CA Management",
}

// NewAdmin creates a new admin view.
func NewAdmin(client *api.Client) Admin {
	cols := []components.Column{
		{Title: "Username", Width: 20},
		{Title: "Display Name", Width: 25},
		{Title: "Email", Width: 30},
		{Title: "Role", Width: 12},
	}
	return Admin{
		client:  client,
		table:   components.NewTable(cols, 15),
		page:    1,
		spinner: components.NewSpinner(),
	}
}

// SetSize updates dimensions.
func (a *Admin) SetSize(width, height int) {
	a.width = width
	a.height = height
	a.table.SetSize(width, height-5)
}

// Init initializes.
func (a *Admin) Init() tea.Cmd { return nil }

// Update handles messages.
func (a *Admin) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case AdminDataMsg:
		a.spinner.Stop()
		if msg.Err != nil {
			a.err = msg.Err.Error()
			return nil
		}
		a.users = msg.Users
		a.cas = msg.CAs
		a.total = msg.Total
		if a.mode == AdminModeUsers {
			a.table.SetRows(a.buildUserRows())
		} else {
			a.table.SetRows(a.buildCARows())
		}
		return nil

	case components.ClearToastMsg:
		a.toast.Hide()
		return nil

	case tea.KeyMsg:
		switch a.mode {
		case AdminModeMenu:
			switch msg.String() {
			case "up", "k":
				if a.menuIdx > 0 {
					a.menuIdx--
				}
			case "down", "j":
				if a.menuIdx < len(adminMenuItems)-1 {
					a.menuIdx++
				}
			case "enter":
				a.mode = AdminMode(a.menuIdx + 1)
				a.page = 1
				cmd := a.spinner.Start("Loading...")
				return tea.Batch(cmd, a.load())
			}
		default:
			switch msg.String() {
			case "esc":
				a.mode = AdminModeMenu
			case "r", "f5":
				cmd := a.spinner.Start("Refreshing...")
				return tea.Batch(cmd, a.load())
			default:
				return a.table.Update(msg)
			}
		}
	}
	return a.spinner.Update(msg)
}

func (a *Admin) load() tea.Cmd {
	mode := a.mode
	page := a.page
	return func() tea.Msg {
		ctx := context.Background()
		switch mode {
		case AdminModeUsers:
			users, err := a.client.ListAdminUsers(ctx, page, 20)
			if err != nil {
				return AdminDataMsg{Err: err}
			}
			return AdminDataMsg{Users: users.List, Total: users.Total}
		case AdminModeCAs:
			cas, err := a.client.ListAdminCAs(ctx, page, 20)
			if err != nil {
				return AdminDataMsg{Err: err}
			}
			return AdminDataMsg{CAs: cas.List, Total: cas.Total}
		}
		return AdminDataMsg{}
	}
}

func (a *Admin) buildUserRows() []components.Row {
	rows := make([]components.Row, len(a.users))
	for i, u := range a.users {
		rows[i] = components.Row{
			u.Username,
			u.DisplayName,
			u.Email,
			api.RoleName(u.Role),
		}
	}
	return rows
}

func (a *Admin) buildCARows() []components.Row {
	rows := make([]components.Row, len(a.cas))
	for i, ca := range a.cas {
		avail := "✓"
		if !ca.Available {
			avail = "✗"
		}
		rows[i] = components.Row{
			ca.CN,
			ca.Algorithm,
			ca.NotAfter.Format("2006-01-02"),
			avail,
		}
	}
	return rows
}

// View renders the admin view.
func (a *Admin) View() string {
	var sb strings.Builder
	sb.WriteString(tui.TitleStyle.Render("⚙️ Admin"))
	sb.WriteString("\n\n")

	if a.mode == AdminModeMenu {
		for i, item := range adminMenuItems {
			if i == a.menuIdx {
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

	title := adminMenuItems[a.mode-1]
	sb.WriteString(tui.SubtitleStyle.Render(title))
	sb.WriteString("\n\n")

	if a.spinner.IsActive() {
		sb.WriteString(a.spinner.View())
		return sb.String()
	}
	if a.err != "" {
		sb.WriteString(tui.DangerStyle.Render("Error: " + a.err))
		sb.WriteString("\n")
		sb.WriteString(tui.HelpStyle.Render("Press r to retry • esc: back"))
		return sb.String()
	}

	total := fmt.Sprintf("Total: %d | Page: %d", a.total, a.page)
	sb.WriteString(tui.MutedStyle.Render(total))
	sb.WriteString("\n")
	sb.WriteString(a.table.View())
	sb.WriteString("\n")
	sb.WriteString(tui.HelpStyle.Render("r: refresh • esc: back • PgUp/PgDn: page"))

	if a.toast.IsVisible() {
		sb.WriteString("\n" + a.toast.View())
	}
	return sb.String()
}
