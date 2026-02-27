package components

import (
	"fmt"
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
	Fields     []*FormField
	focused    int
	title      string
	scrollOff  int // index of first visible field
	maxVisible int // max fields to show (0 = show all)
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

// SetHeight configures the maximum number of visible fields based on available height.
// Each field takes ~3 lines (label + input + gap). Title(2) + help(1) are reserved.
func (f *Form) SetHeight(height int) {
if height < 5 {
f.maxVisible = 1
return
}
// Title(2) + per-field(3) + help(1)
available := height - 3
n := available / 3
if n < 1 {
n = 1
}
f.maxVisible = n
f.clampScroll()
}

func (f *Form) clampScroll() {
if f.maxVisible <= 0 || len(f.Fields) <= f.maxVisible {
f.scrollOff = 0
return
}
if f.scrollOff > len(f.Fields)-f.maxVisible {
f.scrollOff = len(f.Fields) - f.maxVisible
}
if f.scrollOff < 0 {
f.scrollOff = 0
}
// Ensure focused field is visible
if f.focused < f.scrollOff {
f.scrollOff = f.focused
} else if f.focused >= f.scrollOff+f.maxVisible {
f.scrollOff = f.focused - f.maxVisible + 1
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
			f.clampScroll()
			return nil
		case "shift+tab", "up":
			f.Fields[f.focused].input.Blur()
			f.focused = (f.focused - 1 + len(f.Fields)) % len(f.Fields)
			f.Fields[f.focused].input.Focus()
			f.clampScroll()
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

	start := f.scrollOff
	end := len(f.Fields)
	if f.maxVisible > 0 && end > start+f.maxVisible {
		end = start + f.maxVisible
	}

	for i := start; i < end; i++ {
		field := f.Fields[i]
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

	// Scroll indicator when not all fields fit
	if f.maxVisible > 0 && len(f.Fields) > f.maxVisible {
		shown := end - start
		total := len(f.Fields)
		info := st.MutedStyle.Render(
			fmt.Sprintf("  Fields %d–%d of %d  (tab to advance)", start+1, start+shown, total),
		)
		sb.WriteString(info)
		sb.WriteString("\n\n")
	}

	sb.WriteString(st.HelpStyle.Render("tab: next field • shift+tab: prev field"))

	return sb.String()
}
