package components

import (
"github.com/charmbracelet/bubbles/spinner"
tea "github.com/charmbracelet/bubbletea"
st "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
)

// Spinner is a loading indicator.
type Spinner struct {
model   spinner.Model
active  bool
message string
}

// NewSpinner creates a new spinner.
func NewSpinner() Spinner {
s := spinner.New()
s.Spinner = spinner.Dot
s.Style = st.TitleStyle
return Spinner{model: s}
}

// Start activates the spinner with a message.
func (s *Spinner) Start(msg string) tea.Cmd {
s.active = true
s.message = msg
return s.model.Tick
}

// Stop deactivates the spinner.
func (s *Spinner) Stop() {
s.active = false
}

// IsActive returns whether the spinner is active.
func (s *Spinner) IsActive() bool {
return s.active
}

// Update processes spinner tick messages.
func (s *Spinner) Update(msg tea.Msg) tea.Cmd {
if !s.active {
return nil
}
var cmd tea.Cmd
s.model, cmd = s.model.Update(msg)
return cmd
}

// View renders the spinner.
func (s *Spinner) View() string {
if !s.active {
return ""
}
return s.model.View() + " " + st.NormalStyle.Render(s.message)
}
