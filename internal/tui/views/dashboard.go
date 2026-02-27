package views

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
	tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/components"
)

// DashboardStats holds quick stats for the dashboard.
type DashboardStats struct {
	CACount  int
	SSLCount int
	Error    string
}

// DashboardStatsMsg carries dashboard stats from API.
type DashboardStatsMsg DashboardStats

// Dashboard is the main dashboard view.
type Dashboard struct {
	client  *api.Client
	profile *api.UserProfile
	stats   DashboardStats
	spinner components.Spinner
	width   int
	height  int
}

// NewDashboard creates a new dashboard view.
func NewDashboard(client *api.Client, profile *api.UserProfile) Dashboard {
	return Dashboard{
		client:  client,
		profile: profile,
		spinner: components.NewSpinner(),
	}
}

// SetSize updates the view dimensions.
func (d *Dashboard) SetSize(width, height int) {
	d.width = width
	d.height = height
}

// Init fetches dashboard stats.
func (d *Dashboard) Init() tea.Cmd {
	cmd := d.spinner.Start("Loading stats...")
	return tea.Batch(cmd, d.fetchStats())
}

func (d *Dashboard) fetchStats() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		stats := DashboardStats{}

		// Try to fetch CA count
		cas, err := d.client.ListUserCAs(ctx, 1, 1)
		if err != nil {
			if isUnauthorized(err) {
				return SessionExpiredMsg{}
			}
		} else {
			stats.CACount = int(cas.Total)
		}

		// Try to fetch SSL count
		ssls, err := d.client.ListUserSSLCerts(ctx, 1, 1)
		if err != nil {
			if isUnauthorized(err) {
				return SessionExpiredMsg{}
			}
		} else {
			stats.SSLCount = int(ssls.Total)
		}

		return DashboardStatsMsg(stats)
	}
}

// Update handles messages.
func (d *Dashboard) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case DashboardStatsMsg:
		d.spinner.Stop()
		d.stats = DashboardStats(msg)
		return nil
	}
	return d.spinner.Update(msg)
}

// View renders the dashboard.
func (d *Dashboard) View() string {
	var sb strings.Builder

	sb.WriteString(tui.TitleStyle.Render("📊 Dashboard"))
	sb.WriteString("\n\n")

	if d.profile != nil {
		sb.WriteString(tui.SubtitleStyle.Render(fmt.Sprintf("Welcome, %s!", d.profile.DisplayName)))
		sb.WriteString("\n")
		roleStyled := tui.RoleStyle(d.profile.Role).Render(api.RoleName(d.profile.Role))
		sb.WriteString(tui.NormalStyle.Render("Role: ") + roleStyled + tui.NormalStyle.Render(" | Email: "+d.profile.Email))
		sb.WriteString("\n\n")
	}

	if d.spinner.IsActive() {
		sb.WriteString(d.spinner.View())
		sb.WriteString("\n")
		return sb.String()
	}

	// Stats cards
	sb.WriteString(tui.SubtitleStyle.Render("Quick Stats"))
	sb.WriteString("\n\n")

	caCard := tui.BorderStyle.Render(fmt.Sprintf(
		"%s\n\n%s",
		tui.NormalStyle.Render("🔐 CA Certificates"),
		tui.TitleStyle.Render(fmt.Sprintf("%d", d.stats.CACount)),
	))
	sslCard := tui.BorderStyle.Render(fmt.Sprintf(
		"%s\n\n%s",
		tui.NormalStyle.Render("📜 SSL Certificates"),
		tui.TitleStyle.Render(fmt.Sprintf("%d", d.stats.SSLCount)),
	))

	cards := lipgloss.JoinHorizontal(lipgloss.Top, caCard, "   ", sslCard)
	sb.WriteString(cards)
	sb.WriteString("\n\n")

	// Navigation hints
	sb.WriteString(tui.SubtitleStyle.Render("Navigation"))
	sb.WriteString("\n\n")

	hints := []struct{ key, desc string }{
		{"↑/↓ or j/k", "Navigate sidebar"},
		{"enter", "Select section"},
		{"r", "Refresh"},
		{"?", "Show help"},
		{"q", "Quit"},
	}
	for _, h := range hints {
		sb.WriteString(tui.KeyStyle.Render(h.key))
		sb.WriteString(tui.HelpStyle.Render("  " + h.desc))
		sb.WriteString("\n")
	}

	return sb.String()
}
