package views

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
	tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/components"
)

// ProfileUpdatedMsg is sent after a profile update.
type ProfileUpdatedMsg struct {
	Err error
}

// Profile is the user profile view.
type Profile struct {
	client  *api.Client
	profile *api.UserProfile
	form    components.Form
	fields  []*components.FormField
	spinner components.Spinner
	toast   components.Toast
	err     string
	width   int
	height  int
}

// NewProfile creates a new profile view.
func NewProfile(client *api.Client, profile *api.UserProfile) Profile {
	fields := []*components.FormField{
		{Label: "Display Name", Placeholder: "Display name"},
		{Label: "Email", Placeholder: "Email address"},
		{Label: "Old Password", Placeholder: "Current password (to change password)", EchoMode: textinput.EchoPassword},
		{Label: "New Password", Placeholder: "New password (leave blank to keep)", EchoMode: textinput.EchoPassword},
	}
	f := components.NewForm("👤 Update Profile", fields)
	p := Profile{
		client:  client,
		profile: profile,
		form:    f,
		fields:  fields,
		spinner: components.NewSpinner(),
	}
	if profile != nil {
		f.SetValue(0, profile.DisplayName)
		f.SetValue(1, profile.Email)
	}
	return p
}

// SetSize updates dimensions.
func (p *Profile) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// Init initializes.
func (p *Profile) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages.
func (p *Profile) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if p.spinner.IsActive() {
			return p.spinner.Update(msg)
		}
		if msg.String() == "enter" && p.form.FocusedIndex() == len(p.fields)-1 {
			return p.submit()
		}
	case ProfileUpdatedMsg:
		p.spinner.Stop()
		if msg.Err != nil {
			p.err = msg.Err.Error()
		} else {
			return p.toast.Show("Profile updated successfully!", components.ToastSuccess)
		}
		return nil
	case components.ClearToastMsg:
		p.toast.Hide()
		return nil
	}
	spinCmd := p.spinner.Update(msg)
	formCmd := p.form.Update(msg)
	return tea.Batch(spinCmd, formCmd)
}

func (p *Profile) submit() tea.Cmd {
	p.err = ""
	req := api.UpdateProfileRequest{
		DisplayName: p.form.Value(0),
		Email:       p.form.Value(1),
		OldPassword: p.form.Value(2),
		NewPassword: p.form.Value(3),
	}
	cmd := p.spinner.Start("Updating profile...")
	return tea.Batch(cmd, func() tea.Msg {
		err := p.client.UpdateProfile(context.Background(), req)
		return ProfileUpdatedMsg{Err: err}
	})
}

// View renders the profile view.
func (p *Profile) View() string {
	var sb strings.Builder

	if p.profile != nil {
		sb.WriteString(tui.SubtitleStyle.Render("Current Profile"))
		sb.WriteString("\n")
		sb.WriteString(tui.KeyStyle.Render("Username:"))
		sb.WriteString(" " + tui.NormalStyle.Render(p.profile.Username) + "\n")
		sb.WriteString(tui.KeyStyle.Render("Role:"))
		sb.WriteString(" " + tui.NormalStyle.Render(api.RoleName(p.profile.Role)) + "\n\n")
	}

	sb.WriteString(p.form.View())

	if p.spinner.IsActive() {
		sb.WriteString(p.spinner.View())
		sb.WriteString("\n")
	}
	if p.err != "" {
		sb.WriteString(tui.DangerStyle.Render("✗ " + p.err))
		sb.WriteString("\n")
	}
	if p.toast.IsVisible() {
		sb.WriteString(p.toast.View())
		sb.WriteString("\n")
	}
	sb.WriteString(tui.HelpStyle.Render("tab: next field • enter (on last field): submit • esc: back"))
	return sb.String()
}
