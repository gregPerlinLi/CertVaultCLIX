package views

import (
"context"
"fmt"
"strings"

"github.com/charmbracelet/bubbles/textinput"
tea "github.com/charmbracelet/bubbletea"
"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/components"
)

// SuperadminMode indicates which superadmin sub-view is active.
type SuperadminMode int

const (
SuperadminModeMenu              SuperadminMode = iota // 0 - top-level menu
SuperadminModeSessions                                // 1 - all sessions list
SuperadminModeSessionDetail                           // 2 - session detail (shared context)
SuperadminModeUsers                                   // 3 - user list
SuperadminModeUserCreate                              // 4 - create-user form
SuperadminModeUserDetail                              // 5 - user detail menu
SuperadminModeUserEdit                                // 6 - edit user info form
SuperadminModeUserRoleChange                          // 7 - change role selector
SuperadminModeUserPasswordChange                      // 8 - change password form
SuperadminModeUserSessions                            // 9 - user sessions list
)

// superadminDialogAction describes what a confirmed dialog should do.
type superadminDialogAction int

const (
saDialogForceLogoutByUsername superadminDialogAction = iota
saDialogDeleteUser
)

// SuperadminDataMsg carries data loaded from the API for sessions/users lists.
type SuperadminDataMsg struct {
Sessions []api.LoginRecord
Users    []api.AdminUser
Total    int64
Err      error
}

// SuperadminUserOpMsg carries the result of a user write-operation.
type SuperadminUserOpMsg struct {
Err error
}

// SuperadminUserSessionsMsg carries user-specific session data.
type SuperadminUserSessionsMsg struct {
Sessions []api.LoginRecord
Total    int64
Err      error
}

// saRefreshedUserMsg is an internal message used after a user operation to
// refresh both the list and the selectedUser pointer.
type saRefreshedUserMsg struct {
users     []api.AdminUser
total     int64
refreshed *api.AdminUser
}

var superadminMenuItems = []string{
"All Sessions",
"User Management",
}

var userDetailMenuItems = []string{
"Edit User Info",
"Change Role",
"Change Password",
"Delete User",
"View Sessions",
}

// saRoleNames are the human-readable role names in selector order.
var saRoleNames = []string{"User (1)", "Admin (2)", "Superadmin (3)"}

// saRoleValues maps selector index -> API role integer.
var saRoleValues = []int{1, 2, 3}

func saRoleValueToIdx(v int) int {
for i, rv := range saRoleValues {
if rv == v {
return i
}
}
return 0
}

// Superadmin is the superadmin management view.
type Superadmin struct {
client  *api.Client
mode    SuperadminMode
menuIdx int
table   components.Table

// Sessions data
sessions              []api.LoginRecord
selectedSession       *api.LoginRecord
sessionDetailPrevMode SuperadminMode // mode to return to from session detail

// Users data
users             []api.AdminUser
selectedUser      *api.AdminUser
userDetailMenuIdx int

// Create-user form
createUserForm    components.Form
createUserFields  []*components.FormField
createUserRoleIdx int // 0=User,1=Admin,2=Superadmin

// Edit-user form
editUserForm   components.Form
editUserFields []*components.FormField

// Password-change form
passwordForm   components.Form
passwordFields []*components.FormField

// Role-change selector
roleIdx int // index into saRoleValues

// User-sessions data
userSessions      []api.LoginRecord
userSessionsTotal int64
userSessionsPage  int

// Common
total        int64
page         int
spinner      components.Spinner
toast        components.Toast
dialog       *components.Dialog
dialogAction superadminDialogAction
err          string
width        int
height       int
}

// NewSuperadmin creates a new superadmin view.
func NewSuperadmin(client *api.Client) Superadmin {
cols := []components.Column{
{Title: "Username", Width: 20},
{Title: "IP Address", Width: 18},
{Title: "Browser", Width: 22},
{Title: "Login At", Width: 22},
}

// Create-user form
cuFields := []*components.FormField{
{Label: "Username (required)", Placeholder: "e.g. johndoe"},
{Label: "Display Name", Placeholder: "e.g. John Doe"},
{Label: "Email", Placeholder: "e.g. john@example.com"},
{Label: "Password (required)", Placeholder: "", EchoMode: textinput.EchoPassword},
{Label: "Role (use up/down to select)", Placeholder: ""},
}
cuForm := components.NewForm("Plus Create User", cuFields)
cuForm.SetValue(4, saRoleNames[0])

// Edit-user form
euFields := []*components.FormField{
{Label: "Display Name", Placeholder: ""},
{Label: "Email", Placeholder: ""},
}
euForm := components.NewForm("Edit User Info", euFields)

// Password-change form
pwFields := []*components.FormField{
{Label: "New Password (required)", Placeholder: "", EchoMode: textinput.EchoPassword},
}
pwForm := components.NewForm("Change Password", pwFields)

return Superadmin{
client:           client,
table:            components.NewTable(cols, 15),
page:             1,
userSessionsPage: 1,
spinner:          components.NewSpinner(),
createUserForm:   cuForm,
createUserFields: cuFields,
editUserForm:     euForm,
editUserFields:   euFields,
passwordForm:     pwForm,
passwordFields:   pwFields,
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
if isUnauthorized(msg.Err) {
return func() tea.Msg { return SessionExpiredMsg{} }
}
s.err = msg.Err.Error()
return nil
}
s.err = ""
s.sessions = msg.Sessions
s.users = msg.Users
s.total = msg.Total
if s.mode == SuperadminModeSessions {
s.table.SetRows(s.buildSessionRows())
} else {
s.table.SetRows(s.buildUserRows())
}
return nil

case saRefreshedUserMsg:
s.spinner.Stop()
s.err = ""
s.users = msg.users
s.total = msg.total
if msg.refreshed != nil {
s.selectedUser = msg.refreshed
}
return nil

case SuperadminUserOpMsg:
s.spinner.Stop()
if msg.Err != nil {
if isUnauthorized(msg.Err) {
return func() tea.Msg { return SessionExpiredMsg{} }
}
s.err = msg.Err.Error()
return nil
}
s.err = ""
s.toast.Show("Done", 2)
// After user op: decide where to go.
switch s.mode {
case SuperadminModeUserCreate:
s.mode = SuperadminModeUsers
s.page = 1
cmd := s.spinner.Start("Loading...")
return tea.Batch(cmd, s.load())
case SuperadminModeUserEdit, SuperadminModeUserRoleChange, SuperadminModeUserPasswordChange:
s.mode = SuperadminModeUserDetail
cmd := s.spinner.Start("Refreshing...")
return tea.Batch(cmd, s.refreshSelectedUser())
case SuperadminModeUserSessions:
// Force-logout succeeded; reload user sessions.
cmd := s.spinner.Start("Reloading...")
return tea.Batch(cmd, s.loadUserSessions())
}
return nil

case SuperadminUserSessionsMsg:
s.spinner.Stop()
if msg.Err != nil {
if isUnauthorized(msg.Err) {
return func() tea.Msg { return SessionExpiredMsg{} }
}
s.err = msg.Err.Error()
return nil
}
s.err = ""
s.userSessions = msg.Sessions
s.userSessionsTotal = msg.Total
s.table.SetRows(s.buildUserSessionRows())
return nil

case components.ConfirmMsg:
s.dialog = nil
if msg.Confirmed {
switch s.dialogAction {
case saDialogForceLogoutByUsername:
return s.forceLogoutByUsername()
case saDialogDeleteUser:
return s.deleteUser()
}
}
return nil

case components.ClearToastMsg:
s.toast.Hide()
return nil

case tea.MouseMsg:
if s.mode == SuperadminModeSessions || s.mode == SuperadminModeUsers || s.mode == SuperadminModeUserSessions {
return s.table.Update(msg)
}

case tea.KeyMsg:
if s.dialog != nil {
cmd, _ := s.dialog.Update(msg)
return cmd
}

switch s.mode {

// Main menu
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
s.err = ""
s.page = 1
if s.menuIdx == 0 {
s.mode = SuperadminModeSessions
s.table.Columns = []components.Column{
{Title: "Username", Width: 20},
{Title: "IP Address", Width: 18},
{Title: "Browser", Width: 22},
{Title: "Login At", Width: 22},
}
} else {
s.mode = SuperadminModeUsers
s.table.Columns = []components.Column{
{Title: "Username", Width: 20},
{Title: "Display Name", Width: 25},
{Title: "Email", Width: 28},
{Title: "Role", Width: 12},
}
}
cmd := s.spinner.Start("Loading...")
return tea.Batch(cmd, s.load())
}

// All sessions list
case SuperadminModeSessions:
switch msg.String() {
case "esc":
s.mode = SuperadminModeMenu
case "enter":
idx := s.table.SelectedIndex()
if idx >= 0 && idx < len(s.sessions) {
s.selectedSession = &s.sessions[idx]
s.sessionDetailPrevMode = SuperadminModeSessions
s.mode = SuperadminModeSessionDetail
}
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
idx := s.table.SelectedIndex()
if idx >= 0 && idx < len(s.sessions) {
s.dialogAction = saDialogForceLogoutByUsername
d := components.NewDialog("Force Logout",
fmt.Sprintf("Force logout all sessions for %s?", s.sessions[idx].Username))
s.dialog = &d
}
default:
return s.table.Update(msg)
}

// Session detail (used for both all-sessions and user-sessions contexts)
case SuperadminModeSessionDetail:
if msg.String() == "esc" {
s.mode = s.sessionDetailPrevMode
}

// User list
case SuperadminModeUsers:
switch msg.String() {
case "esc":
s.mode = SuperadminModeMenu
case "enter":
idx := s.table.SelectedIndex()
if idx >= 0 && idx < len(s.users) {
u := s.users[idx]
s.selectedUser = &u
s.userDetailMenuIdx = 0
s.mode = SuperadminModeUserDetail
}
case "n":
s.err = ""
s.createUserForm.Reset()
s.createUserRoleIdx = 0
s.createUserForm.SetValue(4, saRoleNames[0])
s.mode = SuperadminModeUserCreate
return textinput.Blink
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
idx := s.table.SelectedIndex()
if idx >= 0 && idx < len(s.users) {
u := s.users[idx]
s.selectedUser = &u
s.dialogAction = saDialogDeleteUser
d := components.NewDialog("Delete User",
fmt.Sprintf("Delete user %s? This cannot be undone.", u.Username))
s.dialog = &d
}
default:
return s.table.Update(msg)
}

// Create-user form
case SuperadminModeUserCreate:
switch msg.String() {
case "esc":
s.err = ""
s.mode = SuperadminModeUsers
case "enter":
if s.createUserForm.FocusedIndex() == len(s.createUserFields)-1 {
return s.submitCreateUser()
}
return s.createUserForm.Update(msg)
default:
// Intercept up/down on the role field (index 4).
if s.createUserForm.FocusedIndex() == 4 {
switch msg.String() {
case "up", "k":
if s.createUserRoleIdx > 0 {
s.createUserRoleIdx--
}
s.createUserForm.SetValue(4, saRoleNames[s.createUserRoleIdx])
return nil
case "down", "j":
if s.createUserRoleIdx < len(saRoleNames)-1 {
s.createUserRoleIdx++
}
s.createUserForm.SetValue(4, saRoleNames[s.createUserRoleIdx])
return nil
case "tab", "shift+tab":
return s.createUserForm.Update(msg)
}
return nil
}
return s.createUserForm.Update(msg)
}

// User detail menu
case SuperadminModeUserDetail:
switch msg.String() {
case "esc":
s.mode = SuperadminModeUsers
case "up", "k":
if s.userDetailMenuIdx > 0 {
s.userDetailMenuIdx--
}
case "down", "j":
if s.userDetailMenuIdx < len(userDetailMenuItems)-1 {
s.userDetailMenuIdx++
}
case "enter":
s.err = ""
switch s.userDetailMenuIdx {
case 0: // Edit User Info
s.editUserForm.Reset()
if s.selectedUser != nil {
s.editUserForm.SetValue(0, s.selectedUser.DisplayName)
s.editUserForm.SetValue(1, s.selectedUser.Email)
}
s.mode = SuperadminModeUserEdit
return textinput.Blink
case 1: // Change Role
if s.selectedUser != nil {
s.roleIdx = saRoleValueToIdx(s.selectedUser.Role)
}
s.mode = SuperadminModeUserRoleChange
case 2: // Change Password
s.passwordForm.Reset()
s.mode = SuperadminModeUserPasswordChange
return textinput.Blink
case 3: // Delete User
if s.selectedUser != nil {
s.dialogAction = saDialogDeleteUser
d := components.NewDialog("Delete User",
fmt.Sprintf("Delete user %s? This cannot be undone.", s.selectedUser.Username))
s.dialog = &d
}
case 4: // View Sessions
if s.selectedUser != nil {
s.userSessionsPage = 1
s.table.Columns = []components.Column{
{Title: "UUID", Width: 36},
{Title: "IP Address", Width: 18},
{Title: "Browser", Width: 22},
{Title: "Login At", Width: 20},
{Title: "Online", Width: 7},
}
s.mode = SuperadminModeUserSessions
cmd := s.spinner.Start("Loading sessions...")
return tea.Batch(cmd, s.loadUserSessions())
}
}
}

// Edit-user form
case SuperadminModeUserEdit:
switch msg.String() {
case "esc":
s.err = ""
s.mode = SuperadminModeUserDetail
case "enter":
if s.editUserForm.FocusedIndex() == len(s.editUserFields)-1 {
return s.submitEditUser()
}
return s.editUserForm.Update(msg)
default:
return s.editUserForm.Update(msg)
}

// Role-change selector
case SuperadminModeUserRoleChange:
switch msg.String() {
case "esc":
s.err = ""
s.mode = SuperadminModeUserDetail
case "up", "k":
if s.roleIdx > 0 {
s.roleIdx--
}
case "down", "j":
if s.roleIdx < len(saRoleNames)-1 {
s.roleIdx++
}
case "enter":
return s.submitChangeRole()
}

// Password-change form
case SuperadminModeUserPasswordChange:
switch msg.String() {
case "esc":
s.err = ""
s.mode = SuperadminModeUserDetail
case "enter":
return s.submitChangePassword()
default:
return s.passwordForm.Update(msg)
}

// User sessions list
case SuperadminModeUserSessions:
switch msg.String() {
case "esc":
s.mode = SuperadminModeUserDetail
case "enter":
idx := s.table.SelectedIndex()
if idx >= 0 && idx < len(s.userSessions) {
s.selectedSession = &s.userSessions[idx]
s.sessionDetailPrevMode = SuperadminModeUserSessions
s.mode = SuperadminModeSessionDetail
}
case "r", "f5":
cmd := s.spinner.Start("Refreshing...")
return tea.Batch(cmd, s.loadUserSessions())
case "[":
if s.userSessionsPage > 1 {
s.userSessionsPage--
return s.loadUserSessions()
}
case "]":
if int64(s.userSessionsPage*20) < s.userSessionsTotal {
s.userSessionsPage++
return s.loadUserSessions()
}
case "d", "delete":
if s.selectedUser != nil {
s.dialogAction = saDialogForceLogoutByUsername
d := components.NewDialog("Force Logout",
fmt.Sprintf("Force logout all sessions for user %s?", s.selectedUser.Username))
s.dialog = &d
}
default:
return s.table.Update(msg)
}
}
}
return s.spinner.Update(msg)
}

// -- Data loading ------------------------------------------------------------

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

func (s *Superadmin) loadUserSessions() tea.Cmd {
if s.selectedUser == nil {
return nil
}
username := s.selectedUser.Username
page := s.userSessionsPage
return func() tea.Msg {
sessions, err := s.client.ListUserSessionsBySuperadmin(context.Background(), username, page, 20)
if err != nil {
return SuperadminUserSessionsMsg{Err: err}
}
return SuperadminUserSessionsMsg{Sessions: sessions.List, Total: sessions.Total}
}
}

// refreshSelectedUser reloads the user list and looks up the selected user by username.
func (s *Superadmin) refreshSelectedUser() tea.Cmd {
username := ""
if s.selectedUser != nil {
username = s.selectedUser.Username
}
page := s.page
return func() tea.Msg {
users, err := s.client.ListAdminUsers(context.Background(), page, 20)
if err != nil {
return SuperadminDataMsg{Err: err}
}
for i := range users.List {
if users.List[i].Username == username {
return saRefreshedUserMsg{users: users.List, total: users.Total, refreshed: &users.List[i]}
}
}
return SuperadminDataMsg{Users: users.List, Total: users.Total}
}
}

// -- Write operations --------------------------------------------------------

func (s *Superadmin) forceLogoutByUsername() tea.Cmd {
var username string
switch s.mode {
case SuperadminModeSessions:
idx := s.table.SelectedIndex()
if idx < 0 || idx >= len(s.sessions) {
return nil
}
username = s.sessions[idx].Username
case SuperadminModeUsers, SuperadminModeUserSessions:
if s.selectedUser == nil {
return nil
}
username = s.selectedUser.Username
}
if username == "" {
return nil
}
mode := s.mode
page := s.page
return func() tea.Msg {
err := s.client.ForceLogoutUser(context.Background(), username)
if err != nil {
return SuperadminDataMsg{Err: err}
}
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
case SuperadminModeUserSessions:
return SuperadminUserOpMsg{}
}
return SuperadminDataMsg{}
}
}

func (s *Superadmin) deleteUser() tea.Cmd {
if s.selectedUser == nil {
return nil
}
username := s.selectedUser.Username
page := s.page
return func() tea.Msg {
err := s.client.DeleteSuperadminUser(context.Background(), username)
if err != nil {
return SuperadminDataMsg{Err: err}
}
users, err := s.client.ListAdminUsers(context.Background(), page, 20)
if err != nil {
return SuperadminDataMsg{Err: err}
}
return SuperadminDataMsg{Users: users.List, Total: users.Total}
}
}

func (s *Superadmin) submitCreateUser() tea.Cmd {
username := s.createUserForm.Value(0)
displayName := s.createUserForm.Value(1)
email := s.createUserForm.Value(2)
password := s.createUserForm.Value(3)

if username == "" || password == "" {
s.err = "Username and password are required"
return nil
}
role := saRoleValues[s.createUserRoleIdx]
req := api.CreateUserRequest{
Username:    username,
DisplayName: displayName,
Email:       email,
Password:    password,
Role:        role,
}
cmd := s.spinner.Start("Creating user...")
return tea.Batch(cmd, func() tea.Msg {
_, err := s.client.CreateUser(context.Background(), req)
return SuperadminUserOpMsg{Err: err}
})
}

func (s *Superadmin) submitEditUser() tea.Cmd {
if s.selectedUser == nil {
return nil
}
username := s.selectedUser.Username
req := api.UpdateSuperadminUserRequest{
DisplayName: s.editUserForm.Value(0),
Email:       s.editUserForm.Value(1),
}
cmd := s.spinner.Start("Updating user...")
return tea.Batch(cmd, func() tea.Msg {
err := s.client.UpdateSuperadminUser(context.Background(), username, req)
return SuperadminUserOpMsg{Err: err}
})
}

func (s *Superadmin) submitChangeRole() tea.Cmd {
if s.selectedUser == nil {
return nil
}
username := s.selectedUser.Username
newRole := saRoleValues[s.roleIdx]
req := api.UpdateUserRoleRequest{Username: username, Role: newRole}
cmd := s.spinner.Start("Updating role...")
return tea.Batch(cmd, func() tea.Msg {
err := s.client.UpdateUserRole(context.Background(), req)
return SuperadminUserOpMsg{Err: err}
})
}

func (s *Superadmin) submitChangePassword() tea.Cmd {
if s.selectedUser == nil {
return nil
}
newPassword := s.passwordForm.Value(0)
if newPassword == "" {
s.err = "Password cannot be empty"
return nil
}
username := s.selectedUser.Username
req := api.UpdateSuperadminUserRequest{Password: newPassword}
cmd := s.spinner.Start("Changing password...")
return tea.Batch(cmd, func() tea.Msg {
err := s.client.UpdateSuperadminUser(context.Background(), username, req)
return SuperadminUserOpMsg{Err: err}
})
}

// -- Row builders ------------------------------------------------------------

func (s *Superadmin) buildSessionRows() []components.Row {
rows := make([]components.Row, len(s.sessions))
for i, sess := range s.sessions {
rows[i] = components.Row{
sess.Username,
sess.IPAddress,
truncate(sess.Browser, 30),
formatNotAfter(sess.LoginTime),
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
tui.RoleStyle(u.Role).Render(api.RoleName(u.Role)),
}
}
return rows
}

func (s *Superadmin) buildUserSessionRows() []components.Row {
rows := make([]components.Row, len(s.userSessions))
for i, sess := range s.userSessions {
online := ""
if sess.IsOnline {
online = "Y"
}
browser := sess.Browser
if sess.OS != "" {
browser = sess.Browser + " / " + sess.OS
}
rows[i] = components.Row{
sess.UUID,
sess.IPAddress,
truncate(browser, 22),
formatNotAfter(sess.LoginTime),
online,
}
}
return rows
}

// -- View rendering ----------------------------------------------------------

// View renders the superadmin view.
func (s *Superadmin) View() string {
var sb strings.Builder
sb.WriteString(tui.TitleStyle.Render("Superadmin"))
sb.WriteString("\n\n")

if s.dialog != nil {
return sb.String() + s.dialog.View(s.width)
}

switch s.mode {

case SuperadminModeMenu:
for i, item := range superadminMenuItems {
if i == s.menuIdx {
sb.WriteString(tui.SelectedStyle.Render("> " + item))
} else {
sb.WriteString(tui.NormalStyle.Render("  " + item))
}
sb.WriteString("\n")
}
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("up/down: select  enter: open"))

case SuperadminModeSessions:
s.renderTableView(&sb, "All Sessions",
"enter: detail  d: force logout  r: refresh  esc: back  [/]: prev/next page")

case SuperadminModeSessionDetail:
s.renderSessionDetailView(&sb)

case SuperadminModeUsers:
s.renderTableView(&sb, "User Management",
"enter: detail  n: new user  d: delete  r: refresh  esc: back  [/]: prev/next page")

case SuperadminModeUserCreate:
sb.WriteString(s.createUserForm.View())
if s.spinner.IsActive() {
sb.WriteString("\n")
sb.WriteString(s.spinner.View())
}
if s.err != "" {
sb.WriteString("\n")
sb.WriteString(tui.DangerStyle.Render("Error: " + s.err))
}
sb.WriteString("\n")
if s.createUserForm.FocusedIndex() == 4 {
sb.WriteString(tui.HelpStyle.Render("up/down: select role  tab: next  enter (last): submit  esc: cancel"))
} else {
sb.WriteString(tui.HelpStyle.Render("tab/down: next  shift+tab/up: prev  enter (last): submit  esc: cancel"))
}

case SuperadminModeUserDetail:
s.renderUserDetailView(&sb)

case SuperadminModeUserEdit:
sb.WriteString(s.editUserForm.View())
if s.spinner.IsActive() {
sb.WriteString("\n")
sb.WriteString(s.spinner.View())
}
if s.err != "" {
sb.WriteString("\n")
sb.WriteString(tui.DangerStyle.Render("Error: " + s.err))
}
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("tab/down: next  shift+tab/up: prev  enter (last): submit  esc: cancel"))

case SuperadminModeUserRoleChange:
s.renderRoleChangeView(&sb)

case SuperadminModeUserPasswordChange:
sb.WriteString(s.passwordForm.View())
if s.spinner.IsActive() {
sb.WriteString("\n")
sb.WriteString(s.spinner.View())
}
if s.err != "" {
sb.WriteString("\n")
sb.WriteString(tui.DangerStyle.Render("Error: " + s.err))
}
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("enter: submit  esc: cancel"))

case SuperadminModeUserSessions:
title := "User Sessions"
if s.selectedUser != nil {
title = fmt.Sprintf("Sessions — %s", s.selectedUser.Username)
}
s.renderTableView(&sb, title,
"enter: detail  d: force logout all  r: refresh  esc: back  [/]: prev/next page")
}

if s.toast.IsVisible() {
sb.WriteString("\n" + s.toast.View())
}
return sb.String()
}

// renderTableView renders a table section with spinner/error/pagination.
func (s *Superadmin) renderTableView(sb *strings.Builder, title, helpText string) {
sb.WriteString(tui.SubtitleStyle.Render(title))
sb.WriteString("\n\n")
if s.spinner.IsActive() {
sb.WriteString(s.spinner.View())
return
}
if s.err != "" {
sb.WriteString(tui.DangerStyle.Render("Error: " + s.err))
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("r: retry  esc: back"))
return
}
var total int64
var page int
if s.mode == SuperadminModeUserSessions {
total = s.userSessionsTotal
page = s.userSessionsPage
} else {
total = s.total
page = s.page
}
sb.WriteString(tui.MutedStyle.Render(fmt.Sprintf("Total: %d | Page: %d", total, page)))
sb.WriteString("\n")
sb.WriteString(s.table.View())
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render(helpText))
}

// renderSessionDetailView renders all fields of the selected session.
func (s *Superadmin) renderSessionDetailView(sb *strings.Builder) {
sb.WriteString(tui.SubtitleStyle.Render("Session Detail"))
sb.WriteString("\n\n")
if s.selectedSession == nil {
sb.WriteString(tui.MutedStyle.Render("No session selected."))
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("esc: back"))
return
}
sess := s.selectedSession
online := "No"
if sess.IsOnline {
online = "Yes"
}
pairs := [][2]string{
{"UUID", sess.UUID},
{"Username", sess.Username},
{"IP Address", sess.IPAddress},
{"Region", sess.Region},
{"Province", sess.Province},
{"City", sess.City},
{"Browser", sess.Browser},
{"OS", sess.OS},
{"Login Time", formatNotAfter(sess.LoginTime)},
{"Online", online},
}
for _, p := range pairs {
label := tui.KeyStyle.Render(fmt.Sprintf("%-12s", p[0]+":"))
sb.WriteString(label + " " + p[1] + "\n")
}
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("esc: back"))
}

// renderUserDetailView renders the selected user info and operation menu.
func (s *Superadmin) renderUserDetailView(sb *strings.Builder) {
if s.selectedUser != nil {
u := s.selectedUser
sb.WriteString(tui.SubtitleStyle.Render("User: " + u.Username))
sb.WriteString("\n\n")
infoFields := [][2]string{
{"Display Name", u.DisplayName},
{"Email", u.Email},
{"Role", tui.RoleStyle(u.Role).Render(api.RoleName(u.Role))},
}
for _, f := range infoFields {
sb.WriteString(tui.NormalStyle.Render(f[0] + ":"))
sb.WriteString("  ")
sb.WriteString(f[1])
sb.WriteString("\n")
}
sb.WriteString("\n")
}
for i, item := range userDetailMenuItems {
if i == s.userDetailMenuIdx {
sb.WriteString(tui.SelectedStyle.Render("> " + item))
} else {
sb.WriteString(tui.NormalStyle.Render("  " + item))
}
sb.WriteString("\n")
}
if s.err != "" {
sb.WriteString("\n")
sb.WriteString(tui.DangerStyle.Render("Error: " + s.err))
sb.WriteString("\n")
}
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("up/down: select  enter: execute  esc: back to users"))
}

// renderRoleChangeView renders the role selector.
func (s *Superadmin) renderRoleChangeView(sb *strings.Builder) {
username := ""
if s.selectedUser != nil {
username = s.selectedUser.Username
}
sb.WriteString(tui.SubtitleStyle.Render("Change Role — " + username))
sb.WriteString("\n\n")
for i, name := range saRoleNames {
if i == s.roleIdx {
sb.WriteString(tui.SelectedStyle.Render("> " + name))
} else {
sb.WriteString(tui.NormalStyle.Render("  " + name))
}
sb.WriteString("\n")
}
if s.err != "" {
sb.WriteString("\n")
sb.WriteString(tui.DangerStyle.Render("Error: " + s.err))
sb.WriteString("\n")
}
sb.WriteString("\n")
sb.WriteString(tui.HelpStyle.Render("up/down: select  enter: confirm  esc: cancel"))
}

// IsAtRoot returns true when the Superadmin view is showing the top-level menu.
func (s *Superadmin) IsAtRoot() bool {
return s.mode == SuperadminModeMenu
}
