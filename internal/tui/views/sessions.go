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

// SessionsLoadedMsg carries session data.
type SessionsLoadedMsg struct {
Sessions []api.LoginRecord
Total    int64
Err      error
}

// Sessions is the session management view.
type Sessions struct {
client   *api.Client
table    components.Table
sessions []api.LoginRecord
total    int64
page     int
spinner  components.Spinner
toast    components.Toast
dialog   *components.Dialog
err      string
width    int
height   int
}

// NewSessions creates a new sessions view.
func NewSessions(client *api.Client) Sessions {
cols := []components.Column{
{Title: "UUID", Width: 36},
{Title: "IP Address", Width: 18},
{Title: "Browser", Width: 22},
{Title: "Login At", Width: 20},
{Title: "Online", Width: 7},
}
return Sessions{
client:  client,
table:   components.NewTable(cols, 15),
page:    1,
spinner: components.NewSpinner(),
}
}

// SetSize updates dimensions.
func (s *Sessions) SetSize(width, height int) {
s.width = width
s.height = height
s.table.SetSize(width, height-5)
}

// Init loads sessions.
func (s *Sessions) Init() tea.Cmd {
cmd := s.spinner.Start("Loading sessions...")
return tea.Batch(cmd, s.load())
}

func (s *Sessions) load() tea.Cmd {
return func() tea.Msg {
sessions, err := s.client.ListUserSessions(context.Background(), s.page, 20)
if err != nil {
return SessionsLoadedMsg{Err: err}
}
return SessionsLoadedMsg{Sessions: sessions.List, Total: sessions.Total}
}
}

// Update handles messages.
func (s *Sessions) Update(msg tea.Msg) tea.Cmd {
switch msg := msg.(type) {
case SessionsLoadedMsg:
s.spinner.Stop()
if msg.Err != nil {
if isUnauthorized(msg.Err) {
return func() tea.Msg { return SessionExpiredMsg{} }
}
s.err = msg.Err.Error()
return nil
}
s.sessions = msg.Sessions
s.total = msg.Total
s.table.SetRows(s.buildRows())
return nil

case components.ConfirmMsg:
s.dialog = nil
if msg.Confirmed {
return s.logoutSelected()
}
return nil

case components.ClearToastMsg:
s.toast.Hide()
return nil

case tea.MouseMsg:
return s.table.Update(msg)

case tea.KeyMsg:
if s.dialog != nil {
cmd, _ := s.dialog.Update(msg)
return cmd
}
switch msg.String() {
case "r", "f5":
cmd := s.spinner.Start("Refreshing...")
return tea.Batch(cmd, s.load())
case "[":
if s.page > 1 {
s.page--
return s.load()
}
case "]":
if int64(s.page*20) < s.total {
s.page++
return s.load()
}
case "d", "delete":
sel := s.table.SelectedIndex()
if sel >= 0 && sel < len(s.sessions) {
d := components.NewDialog("Logout Session",
fmt.Sprintf("Logout session %s?", s.sessions[sel].UUID))
s.dialog = &d
}
return nil
default:
return s.table.Update(msg)
}
}
return s.spinner.Update(msg)
}

func (s *Sessions) logoutSelected() tea.Cmd {
idx := s.table.SelectedIndex()
if idx < 0 || idx >= len(s.sessions) {
return nil
}
uuid := s.sessions[idx].UUID
return func() tea.Msg {
err := s.client.LogoutSession(context.Background(), uuid)
if err != nil {
return SessionsLoadedMsg{Err: err}
}
sessions, err := s.client.ListUserSessions(context.Background(), s.page, 20)
if err != nil {
return SessionsLoadedMsg{Err: err}
}
return SessionsLoadedMsg{Sessions: sessions.List, Total: sessions.Total}
}
}

func (s *Sessions) buildRows() []components.Row {
rows := make([]components.Row, len(s.sessions))
for i, sess := range s.sessions {
online := ""
if sess.IsOnline {
online = "✓"
}
browser := sess.Browser
if sess.OS != "" {
browser = sess.Browser + " / " + sess.OS
}
loginTime := formatNotAfter(sess.LoginTime)
rows[i] = components.Row{
sess.UUID,
sess.IPAddress,
truncate(browser, 22),
loginTime,
online,
}
}
return rows
}

// View renders the sessions view.
func (s *Sessions) View() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("📋 Sessions"))
sb.WriteString("\n\n")

if s.dialog != nil {
return sb.String() + s.dialog.View(s.width)
}

if s.spinner.IsActive() {
sb.WriteString(s.spinner.View())
return sb.String()
}
if s.err != "" {
sb.WriteString(tui.DangerStyle.Render("Error: " + s.err))
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("Press r to retry"))
return sb.String()
}

total := fmt.Sprintf("Total: %d | Page: %d", s.total, s.page)
sb.WriteString(tui.MutedStyle.Render(total))
sb.WriteString("\n")
sb.WriteString(s.table.View())
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("d: logout session • r: refresh • [/]: prev/next page"))

if s.toast.IsVisible() {
sb.WriteString("\n" + s.toast.View())
}
return sb.String()
}

func truncate(s string, n int) string {
if len(s) <= n {
return s
}
return s[:n-3] + "..."
}
