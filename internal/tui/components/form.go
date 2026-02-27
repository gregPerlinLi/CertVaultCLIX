package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	st "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
)

// FormField represents a single form field.
type FormField struct {
	Label       string
	Placeholder string
	EchoMode    textinput.EchoMode
	input       textinput.Model
}

// Form is a multi-field form component.
type Form struct {
	Fields  []*FormField
	focused int
	title   string
}

// NewForm creates a new form with the given fields.
func NewForm(title string, fields []*FormField) Form {
	for i, f := range fields {
		ti := textinput.New()
		ti.Placeholder = f.Placeholder
		ti.EchoMode = f.EchoMode
		ti.CharLimit = 256
		if i == 0 {
			ti.Focus()
		}
		f.input = ti
	}
	return Form{
		Fields: fields,
		title:  title,
	}
}

// Value returns the value of the field at index i.
func (f *Form) Value(i int) string {
	if i < 0 || i >= len(f.Fields) {
		return ""
	}
	return f.Fields[i].input.Value()
}

// SetValue sets the value of the field at index i.
func (f *Form) SetValue(i int, val string) {
	if i < 0 || i >= len(f.Fields) {
		return
	}
	f.Fields[i].input.SetValue(val)
}

// Reset clears all field values.
func (f *Form) Reset() {
	for _, field := range f.Fields {
		field.input.SetValue("")
		field.input.Blur()
	}
	f.focused = 0
	if len(f.Fields) > 0 {
		f.Fields[0].input.Focus()
	}
}

// FocusedIndex returns the currently focused field index.
func (f *Form) FocusedIndex() int {
	return f.focused
}

// Update handles keyboard input for the form.
func (f *Form) Update(msg tea.Msg) tea.Cmd {
	if len(f.Fields) == 0 {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			f.Fields[f.focused].input.Blur()
			f.focused = (f.focused + 1) % len(f.Fields)
			f.Fields[f.focused].input.Focus()
			return nil
		case "shift+tab", "up":
			f.Fields[f.focused].input.Blur()
			f.focused = (f.focused - 1 + len(f.Fields)) % len(f.Fields)
			f.Fields[f.focused].input.Focus()
			return nil
		}
	}

	var cmd tea.Cmd
	f.Fields[f.focused].input, cmd = f.Fields[f.focused].input.Update(msg)
	return cmd
}

// View renders the form.
func (f *Form) View() string {
	var sb strings.Builder

	if f.title != "" {
		sb.WriteString(st.TitleStyle.Render(f.title))
		sb.WriteString("\n\n")
	}

	for i, field := range f.Fields {
		label := st.NormalStyle.Render(field.Label + ":")
		sb.WriteString(label)
		sb.WriteString("\n")

		if i == f.focused {
			sb.WriteString(st.InputFocusStyle.Render(field.input.View()))
		} else {
			sb.WriteString(st.InputStyle.Render(field.input.View()))
		}
		sb.WriteString("\n\n")
	}

	sb.WriteString(st.HelpStyle.Render("tab: next field • shift+tab: prev field"))

	return sb.String()
}
