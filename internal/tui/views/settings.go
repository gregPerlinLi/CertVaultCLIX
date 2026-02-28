package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/config"
	tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/components"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/version"
)

// ServerURLUpdatedMsg is sent after the server URL is updated.
type ServerURLUpdatedMsg struct {
	URL string
}

// Settings is the settings/about view.
type Settings struct {
	cfg     *config.Config
	editing bool
	input   components.Form
	toast   components.Toast
	fields  []*components.FormField
	width   int
	height  int
}

// NewSettings creates a new settings view.
func NewSettings(cfg *config.Config) Settings {
	fields := []*components.FormField{
		{Label: "Server URL", Placeholder: "http://localhost:1888"},
	}
	f := components.NewForm("", fields)
	if cfg != nil {
		f.SetValue(0, cfg.ServerURL)
	}
	return Settings{
		cfg:    cfg,
		fields: fields,
		input:  f,
	}
}

// SetSize updates dimensions.
func (s *Settings) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// Init does nothing.
func (s *Settings) Init() tea.Cmd { return nil }

// Update handles messages.
func (s *Settings) Update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "e":
			if !s.editing {
				s.editing = true
				return nil, false
			}
		case "esc":
			if s.editing {
				s.editing = false
				if s.cfg != nil {
					s.input.SetValue(0, s.cfg.ServerURL)
				}
				return nil, false
			}
		case "enter":
			if s.editing {
				newURL := s.input.Value(0)
				if newURL != "" && s.cfg != nil {
					s.cfg.ServerURL = newURL
					_ = config.Save(s.cfg)
					s.editing = false
					cmd := s.toast.Show("Server URL updated!", components.ToastSuccess)
					return cmd, true
				}
			}
		}
		if s.editing {
			return s.input.Update(msg), false
		}
	case components.ClearToastMsg:
		s.toast.Hide()
		return nil, false
	}
	return nil, false
}

// View renders the settings view.
func (s *Settings) View() string {
	var sb strings.Builder
	sb.WriteString(tui.TitleStyle.Render("⚡ Settings"))
	sb.WriteString("\n\n")

	// Server URL
	sb.WriteString(tui.SubtitleStyle.Render("Server Configuration"))
	sb.WriteString("\n")
	if s.cfg != nil {
		sb.WriteString(tui.KeyStyle.Render("Server URL:"))
		sb.WriteString(" " + tui.NormalStyle.Render(s.cfg.ServerURL) + "\n")
		sb.WriteString(tui.KeyStyle.Render("Config File:"))
		sb.WriteString(" " + tui.MutedStyle.Render(config.Path()) + "\n\n")
	}

	if s.editing {
		sb.WriteString(s.input.View())
		sb.WriteString(tui.HelpStyle.Render("enter: save • esc: cancel"))
	} else {
		sb.WriteString(tui.HelpStyle.Render("e: edit server URL"))
	}

	sb.WriteString("\n\n")

	// About
	sb.WriteString(tui.SubtitleStyle.Render("About"))
	sb.WriteString("\n")
	sb.WriteString(tui.KeyStyle.Render("CertVaultCLIX"))
	sb.WriteString(" — Interactive TUI for CertVault\n")
	sb.WriteString(tui.KeyStyle.Render("Version:"))
	sb.WriteString(" " + tui.NormalStyle.Render(version.String()) + "\n")
	sb.WriteString(tui.KeyStyle.Render("GitHub:"))
	sb.WriteString(" " + tui.NormalStyle.Render("https://github.com/gregPerlinLi/CertVaultCLIX") + "\n")

	if s.toast.IsVisible() {
		sb.WriteString("\n" + s.toast.View())
	}
	return sb.String()
}
