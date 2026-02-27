package tui

import (
"context"
"fmt"

tea "github.com/charmbracelet/bubbletea"
"github.com/charmbracelet/lipgloss"
"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
"github.com/gregPerlinLi/CertVaultCLIX/internal/config"
"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/components"
"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/views"
)

// ViewID identifies the active view.
type ViewID int

const (
ViewLogin ViewID = iota
ViewDashboard
ViewCAList
ViewCADetail
ViewCertList
ViewCertDetail
ViewCertRequest
ViewProfile
ViewSessions
ViewTools
ViewAdmin
ViewSuperadmin
ViewSettings
)

// App is the main Bubble Tea application model.
type App struct {
client    *api.Client
cfg       *config.Config
profile   *api.UserProfile
view      ViewID
prevView  ViewID
width     int
height    int
ready     bool
sidebar   components.Sidebar
statusBar components.StatusBar
help      components.Help
toast     components.Toast

// Views
loginView      *views.Login
dashboardView  *views.Dashboard
caListView     *views.CAList
caDetailView   *views.CADetail
certListView   *views.CertList
certDetailView *views.CertDetail
certReqView    *views.CertRequest
profileView    *views.Profile
sessionsView   *views.Sessions
toolsView      *views.Tools
adminView      *views.Admin
superadminView *views.Superadmin
settingsView   *views.Settings
}

// NewApp creates a new App.
func NewApp(client *api.Client, cfg *config.Config) *App {
loginView := views.NewLogin(client, cfg)
return &App{
client:    client,
cfg:       cfg,
view:      ViewLogin,
loginView: &loginView,
help:      components.NewHelp(components.DefaultEntries()),
}
}

// Init starts the application.
func (a *App) Init() tea.Cmd {
if a.cfg != nil && a.cfg.Session != "" {
return tea.Batch(a.loginView.Init(), a.tryAutoLogin())
}
return a.loginView.Init()
}

// tryAutoLogin attempts to restore a saved session without re-entering credentials.
func (a *App) tryAutoLogin() tea.Cmd {
return func() tea.Msg {
profile, err := a.client.GetProfile(context.Background())
if err != nil {
return nil // session invalid; stay on login screen
}
return views.LoginSuccessMsg{Profile: profile}
}
}

// Update handles all messages.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
switch msg := msg.(type) {
case tea.WindowSizeMsg:
a.width = msg.Width
a.height = msg.Height
a.ready = true
a.updateSizes()
return a, nil

case tea.KeyMsg:
// Global quit: ctrl+c always; 'q' only when not editing text
if msg.String() == "ctrl+c" {
return a, tea.Quit
}
if msg.String() == "q" {
switch a.view {
case ViewLogin, ViewDashboard, ViewCAList, ViewCADetail, ViewCertList, ViewCertDetail, ViewSessions:
return a, tea.Quit
case ViewTools:
if a.toolsView != nil && a.toolsView.IsAtRoot() {
return a, tea.Quit
}
case ViewAdmin:
if a.adminView != nil && a.adminView.IsAtRoot() {
return a, tea.Quit
}
case ViewSuperadmin:
if a.superadminView != nil && a.superadminView.IsAtRoot() {
return a, tea.Quit
}
}
}
// Toggle help
if msg.String() == "?" {
a.help.Toggle()
return a, nil
}
// If help overlay is visible, close on any key
if a.help.IsVisible() {
a.help.Toggle()
return a, nil
}

case views.LoginSuccessMsg:
a.profile = msg.Profile
return a, a.switchToMain()

case components.ClearToastMsg:
a.toast.Hide()
return a, nil
}

// Delegate to current view
var cmd tea.Cmd
switch a.view {
case ViewLogin:
var done bool
cmd, done = a.loginView.Update(msg)
if done {
// LoginSuccessMsg will be dispatched via the cmd
}
return a, cmd

case ViewDashboard:
if cmd := a.handleSidebar(msg); cmd != nil {
return a, cmd
}
if a.dashboardView != nil {
cmd = a.dashboardView.Update(msg)
}

case ViewCAList:
if key, ok := msg.(tea.KeyMsg); ok {
switch key.String() {
case "esc":
a.view = ViewDashboard
return a, nil
case "enter":
ca := a.caListView.SelectedCA()
if ca != nil {
d := views.NewCADetail(ca)
a.caDetailView = &d
a.prevView = ViewCAList
a.view = ViewCADetail
return a, nil
}
}
}
cmd = a.caListView.Update(msg)

case ViewCADetail:
if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
a.view = a.prevView
return a, nil
}

case ViewCertList:
if key, ok := msg.(tea.KeyMsg); ok {
switch key.String() {
case "esc":
a.view = ViewDashboard
return a, nil
case "enter":
cert := a.certListView.SelectedCert()
if cert != nil {
d := views.NewCertDetail(cert)
a.certDetailView = &d
a.prevView = ViewCertList
a.view = ViewCertDetail
return a, nil
}
case "n":
a.view = ViewCertRequest
return a, a.certReqView.Init()
}
}
cmd = a.certListView.Update(msg)

case ViewCertDetail:
if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
a.view = a.prevView
return a, nil
}

case ViewCertRequest:
if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
a.view = ViewCertList
return a, nil
}
cmd = a.certReqView.Update(msg)
if m, ok := msg.(views.CertRequestedMsg); ok && m.Err == nil {
a.view = ViewCertList
return a, a.certListView.Init()
}

case ViewProfile:
if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
a.view = ViewDashboard
return a, nil
}
cmd = a.profileView.Update(msg)

case ViewSessions:
if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
a.view = ViewDashboard
return a, nil
}
cmd = a.sessionsView.Update(msg)

case ViewTools:
if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
if a.toolsView != nil && a.toolsView.IsAtRoot() {
a.view = ViewDashboard
return a, nil
}
}
if a.toolsView != nil {
cmd = a.toolsView.Update(msg)
}

case ViewAdmin:
if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
if a.adminView != nil && a.adminView.IsAtRoot() {
a.view = ViewDashboard
return a, nil
}
if a.adminView != nil {
cmd = a.adminView.Update(msg)
}
return a, cmd
}
if a.adminView != nil {
cmd = a.adminView.Update(msg)
}

case ViewSuperadmin:
if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
if a.superadminView != nil && a.superadminView.IsAtRoot() {
a.view = ViewDashboard
return a, nil
}
if a.superadminView != nil {
cmd = a.superadminView.Update(msg)
}
return a, cmd
}
if a.superadminView != nil {
cmd = a.superadminView.Update(msg)
}

case ViewSettings:
if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
a.view = ViewDashboard
return a, nil
}
if a.settingsView != nil {
var urlUpdated bool
cmd, urlUpdated = a.settingsView.Update(msg)
if urlUpdated {
a.client.SetBaseURL(a.cfg.ServerURL)
}
}
}

return a, cmd
}

// handleSidebar processes sidebar key events.
func (a *App) handleSidebar(msg tea.Msg) tea.Cmd {
key, ok := msg.(tea.KeyMsg)
if !ok {
return nil
}
switch key.String() {
case "up", "k":
a.sidebar.MoveUp()
return nil
case "down", "j":
a.sidebar.MoveDown()
return nil
case "enter":
return a.navigateToSidebarItem()
}
return nil
}

func (a *App) navigateToSidebarItem() tea.Cmd {
id := a.sidebar.SelectedID()
switch id {
case "dashboard":
a.view = ViewDashboard
case "ca_list":
a.view = ViewCAList
return a.caListView.Init()
case "cert_list":
a.view = ViewCertList
return a.certListView.Init()
case "cert_request":
a.view = ViewCertRequest
return a.certReqView.Init()
case "profile":
a.view = ViewProfile
return a.profileView.Init()
case "sessions":
a.view = ViewSessions
return a.sessionsView.Init()
case "tools":
a.view = ViewTools
return a.toolsView.Init()
case "admin":
a.view = ViewAdmin
return a.adminView.Init()
case "superadmin":
a.view = ViewSuperadmin
return a.superadminView.Init()
case "settings":
a.view = ViewSettings
return a.settingsView.Init()
}
return nil
}

func (a *App) switchToMain() tea.Cmd {
dashView := views.NewDashboard(a.client, a.profile)
a.dashboardView = &dashView

caList := views.NewCAList(a.client)
a.caListView = &caList

certList := views.NewCertList(a.client)
a.certListView = &certList

certReq := views.NewCertRequest(a.client)
a.certReqView = &certReq

profileView := views.NewProfile(a.client, a.profile)
a.profileView = &profileView

sessionsView := views.NewSessions(a.client)
a.sessionsView = &sessionsView

toolsView := views.NewTools(a.client)
a.toolsView = &toolsView

adminView := views.NewAdmin(a.client)
a.adminView = &adminView

superadminView := views.NewSuperadmin(a.client)
a.superadminView = &superadminView

settingsView := views.NewSettings(a.cfg)
a.settingsView = &settingsView

items := a.buildSidebarItems()
a.sidebar = components.NewSidebar(items)

role := "User"
username := ""
email := ""
if a.profile != nil {
role = api.RoleName(a.profile.Role)
username = a.profile.Username
email = a.profile.Email
}
a.statusBar = components.StatusBar{
Username: username,
Role:     role,
Server:   a.client.GetBaseURL(),
Status:   email,
}

a.view = ViewDashboard
a.updateSizes()
return dashView.Init()
}

func (a *App) buildSidebarItems() []components.SidebarItem {
items := []components.SidebarItem{
{Icon: "📊", Label: "Dashboard", ID: "dashboard"},
{Icon: "🔐", Label: "CA Certificates", ID: "ca_list"},
{Icon: "📜", Label: "SSL Certificates", ID: "cert_list"},
{Icon: "➕", Label: "Request Cert", ID: "cert_request"},
{Icon: "👤", Label: "Profile", ID: "profile"},
{Icon: "📋", Label: "Sessions", ID: "sessions"},
{Icon: "🛠", Label: "Tools", ID: "tools"},
}

if a.profile != nil && a.profile.Role >= 2 {
items = append(items, components.SidebarItem{Icon: "🔧", Label: "Admin", ID: "admin"})
}
if a.profile != nil && a.profile.Role >= 3 {
items = append(items, components.SidebarItem{Icon: "👑", Label: "Superadmin", ID: "superadmin"})
}

items = append(items, components.SidebarItem{Icon: "⚡", Label: "Settings", ID: "settings"})
return items
}

func (a *App) updateSizes() {
if !a.ready {
return
}
sidebarWidth := 22
statusBarHeight := 1
contentWidth := a.width - sidebarWidth
contentHeight := a.height - statusBarHeight - 1

a.sidebar.SetSize(sidebarWidth, contentHeight)
a.statusBar.SetSize(a.width)
a.help.SetSize(a.width)

if a.loginView != nil {
a.loginView.SetSize(a.width, a.height)
}
if a.dashboardView != nil {
a.dashboardView.SetSize(contentWidth, contentHeight)
}
if a.caListView != nil {
a.caListView.SetSize(contentWidth, contentHeight)
}
if a.certListView != nil {
a.certListView.SetSize(contentWidth, contentHeight)
}
if a.certReqView != nil {
a.certReqView.SetSize(contentWidth, contentHeight)
}
if a.profileView != nil {
a.profileView.SetSize(contentWidth, contentHeight)
}
if a.sessionsView != nil {
a.sessionsView.SetSize(contentWidth, contentHeight)
}
if a.toolsView != nil {
a.toolsView.SetSize(contentWidth, contentHeight)
}
if a.adminView != nil {
a.adminView.SetSize(contentWidth, contentHeight)
}
if a.superadminView != nil {
a.superadminView.SetSize(contentWidth, contentHeight)
}
if a.settingsView != nil {
a.settingsView.SetSize(contentWidth, contentHeight)
}
}

// View renders the application.
func (a *App) View() string {
if !a.ready {
return "Initializing..."
}

if a.view == ViewLogin {
view := a.loginView.View()
if a.help.IsVisible() {
return view + "\n" + a.help.View()
}
return view
}

sidebarView := a.sidebar.View()
contentView := a.currentContentView()

sidebarWidth := 22
contentWidth := a.width - sidebarWidth
if contentWidth < 1 {
contentWidth = 1
}

content := lipgloss.NewStyle().Width(contentWidth).Render(contentView)
mainArea := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, content)

statusBar := a.statusBar.View()

var footer string
if a.help.IsVisible() {
helpView := a.help.View()
footer = lipgloss.PlaceHorizontal(a.width, lipgloss.Center, helpView)
} else {
footer = HelpStyle.Render("? help  q quit")
}

return fmt.Sprintf("%s\n%s\n%s", mainArea, statusBar, footer)
}

func (a *App) currentContentView() string {
switch a.view {
case ViewDashboard:
if a.dashboardView != nil {
return a.dashboardView.View()
}
case ViewCAList:
if a.caListView != nil {
return a.caListView.View()
}
case ViewCADetail:
if a.caDetailView != nil {
return a.caDetailView.View()
}
case ViewCertList:
if a.certListView != nil {
return a.certListView.View()
}
case ViewCertDetail:
if a.certDetailView != nil {
return a.certDetailView.View()
}
case ViewCertRequest:
if a.certReqView != nil {
return a.certReqView.View()
}
case ViewProfile:
if a.profileView != nil {
return a.profileView.View()
}
case ViewSessions:
if a.sessionsView != nil {
return a.sessionsView.View()
}
case ViewTools:
if a.toolsView != nil {
return a.toolsView.View()
}
case ViewAdmin:
if a.adminView != nil {
return a.adminView.View()
}
case ViewSuperadmin:
if a.superadminView != nil {
return a.superadminView.View()
}
case ViewSettings:
if a.settingsView != nil {
return a.settingsView.View()
}
}
return MutedStyle.Render("No content.")
}

// Client returns the API client.
func (a *App) Client() *api.Client {
return a.client
}
