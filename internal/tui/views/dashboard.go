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
BindedCA        int64
RequestedSSL    int64
TotalUsers      int64 // admin+ only
RequestedCACerts int64 // admin+ only
TotalCACerts    int64 // superadmin only
TotalSSLCerts   int64 // superadmin only
Error           string
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
role := 1
if d.profile != nil {
role = d.profile.Role
}
return func() tea.Msg {
ctx := context.Background()
stats := DashboardStats{}

// All users: binded CA count and requested SSL count.
bindedCA, err := d.client.CountUserCAs(ctx)
if err != nil {
if isUnauthorized(err) {
return SessionExpiredMsg{}
}
} else {
stats.BindedCA = bindedCA
}

requestedSSL, err := d.client.CountUserSSLCerts(ctx)
if err != nil {
if isUnauthorized(err) {
return SessionExpiredMsg{}
}
} else {
stats.RequestedSSL = requestedSSL
}

// Admin+ only stats.
if role >= 2 {
totalUsers, err := d.client.CountAdminUsers(ctx)
if err == nil {
stats.TotalUsers = totalUsers
}
requestedCA, err := d.client.CountAdminCAs(ctx)
if err == nil {
stats.RequestedCACerts = requestedCA
}
}

// Superadmin-only stats.
if role >= 3 {
totalCA, err := d.client.CountAllCAs(ctx)
if err == nil {
stats.TotalCACerts = totalCA
}
totalSSL, err := d.client.CountAllSSLCerts(ctx)
if err == nil {
stats.TotalSSLCerts = totalSSL
}
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

func renderStatCard(icon, label string, val int64) string {
return tui.BorderStyle.Render(fmt.Sprintf(
"%s\n\n%s",
tui.NormalStyle.Render(icon+" "+label),
tui.TitleStyle.Render(fmt.Sprintf("%d", val)),
))
}

// View renders the dashboard.
func (d *Dashboard) View() string {
var sb strings.Builder

sb.WriteString(tui.TitleStyle.Render("📊 Dashboard"))
sb.WriteString("\n\n")

role := 1
if d.profile != nil {
sb.WriteString(tui.SubtitleStyle.Render(fmt.Sprintf("Welcome, %s!", d.profile.DisplayName)))
sb.WriteString("\n")
roleStyled := tui.RoleStyle(d.profile.Role).Render(api.RoleName(d.profile.Role))
sb.WriteString(tui.NormalStyle.Render("Role: ") + roleStyled + tui.NormalStyle.Render(" | Email: "+d.profile.Email))
sb.WriteString("\n\n")
role = d.profile.Role
}

if d.spinner.IsActive() {
sb.WriteString(d.spinner.View())
sb.WriteString("\n")
return sb.String()
}

// --- Stats cards ---
sb.WriteString(tui.SubtitleStyle.Render("Quick Stats"))
sb.WriteString("\n\n")

// Row 1: visible to all
row1 := lipgloss.JoinHorizontal(lipgloss.Top,
renderStatCard("🔐", "Binded CA", d.stats.BindedCA),
"   ",
renderStatCard("📜", "Requested SSL Certs", d.stats.RequestedSSL),
)
sb.WriteString(row1)
sb.WriteString("\n\n")

// Row 2: admin+
if role >= 2 {
row2 := lipgloss.JoinHorizontal(lipgloss.Top,
renderStatCard("👥", "Total Users", d.stats.TotalUsers),
"   ",
renderStatCard("🏛", "Requested CA Certs", d.stats.RequestedCACerts),
)
sb.WriteString(row2)
sb.WriteString("\n\n")
}

// Row 3: superadmin
if role >= 3 {
row3 := lipgloss.JoinHorizontal(lipgloss.Top,
renderStatCard("🔒", "Total CA Certs", d.stats.TotalCACerts),
"   ",
renderStatCard("📋", "Total SSL Certs", d.stats.TotalSSLCerts),
)
sb.WriteString(row3)
sb.WriteString("\n\n")
}

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
