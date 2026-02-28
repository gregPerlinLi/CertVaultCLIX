package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	st "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
)

// PathInput is a text input enhanced with Tab-based filesystem path autocompletion.
// Press Tab to expand the current value to the longest matching prefix; pressing Tab
// again when the common prefix has already been filled cycles through all candidates.
type PathInput struct {
	Input       textinput.Model
	completions []string // current candidate list (nil when no suggestion state)
	compIdx     int      // index of the highlighted candidate (-1 = none)
}

// NewPathInput creates a PathInput with the given placeholder and character limit.
func NewPathInput(placeholder string, charLimit int) PathInput {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = charLimit
	return PathInput{Input: ti, compIdx: -1}
}

// Focus passes focus to the underlying text input.
func (p *PathInput) Focus() { p.Input.Focus() }

// Blur removes focus from the underlying text input.
func (p *PathInput) Blur() { p.Input.Blur() }

// SetValue sets the input value and recomputes completions for the new value.
func (p *PathInput) SetValue(v string) {
	p.Input.SetValue(v)
	p.compIdx = -1
	p.completions = p.computeCompletions(p.expandHome(v))
}

// Value returns the current input value.
func (p *PathInput) Value() string { return p.Input.Value() }

// InputView returns the raw text-input view, suitable for wrapping in a styled border.
func (p *PathInput) InputView() string { return p.Input.View() }

// SuggestionsView returns a rendered list of path completions, or an empty string
// when there are no pending suggestions. A sliding window of maxShow entries is
// used so that the currently selected item (compIdx) is always visible.
func (p *PathInput) SuggestionsView() string {
	if len(p.completions) == 0 {
		return ""
	}
	const maxShow = 8

	// Compute the view window so that compIdx is always inside it.
	viewStart := 0
	if p.compIdx >= maxShow {
		viewStart = p.compIdx - maxShow + 1
	}
	viewEnd := viewStart + maxShow
	if viewEnd > len(p.completions) {
		viewEnd = len(p.completions)
	}

	var sb strings.Builder
	if viewStart > 0 {
		sb.WriteString(st.MutedStyle.Render(fmt.Sprintf("  ↑ %d more above", viewStart)))
		sb.WriteString("\n")
	}
	for i := viewStart; i < viewEnd; i++ {
		name := filepath.Base(p.completions[i])
		if strings.HasSuffix(p.completions[i], string(filepath.Separator)) {
			name += string(filepath.Separator)
		}
		if i == p.compIdx {
			sb.WriteString(st.SelectedStyle.Render("▶ " + name))
		} else {
			sb.WriteString(st.MutedStyle.Render("  " + name))
		}
		sb.WriteString("\n")
	}
	if viewEnd < len(p.completions) {
		sb.WriteString(st.MutedStyle.Render(fmt.Sprintf("  ↓ %d more below", len(p.completions)-viewEnd)))
		sb.WriteString("\n")
	}
	return sb.String()
}

// Update processes key events. Completions are recomputed after every keystroke so
// the list is always visible while the user types. Tab fills the longest common
// prefix of all candidates on the first press, then cycles through them on
// subsequent presses. Any printable / delete key resets the cycle state.
func (p *PathInput) Update(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "tab" {
		// Already cycling — advance to the next candidate.
		if len(p.completions) > 0 && p.compIdx >= 0 {
			p.compIdx = (p.compIdx + 1) % len(p.completions)
			p.Input.SetValue(p.completions[p.compIdx])
			return nil
		}
		// Use the live-computed list; it is always current thanks to the
		// recomputation that happens on every non-Tab keystroke.
		candidates := p.completions // always current
		if len(candidates) == 0 {
			return nil
		}
		if len(candidates) == 1 {
			p.Input.SetValue(candidates[0])
			p.completions = nil
			p.compIdx = -1
			return nil
		}
		// Multiple candidates: fill the longest common prefix, then start cycling.
		common := longestCommonPrefix(candidates)
		p.completions = candidates
		path := p.expandHome(p.Input.Value())
		if common != path {
			p.Input.SetValue(common)
			p.compIdx = -1
			return nil
		}
		// Already at the common prefix — begin cycling through candidates.
		p.compIdx = 0
		p.Input.SetValue(p.completions[0])
		return nil
	}
	// Non-Tab key: reset cycle state, forward to text input, recompute live list.
	p.compIdx = -1
	var cmd tea.Cmd
	p.Input, cmd = p.Input.Update(msg)
	// Recompute completions for the updated value so the list is always visible.
	p.completions = p.computeCompletions(p.expandHome(p.Input.Value()))
	return cmd
}

// computeCompletions returns the filesystem entries whose names begin with the
// base-name component of path. Directory entries get a trailing separator appended.
func (p *PathInput) computeCompletions(path string) []string {
	var dir, prefix string
	if path == "" || strings.HasSuffix(path, string(filepath.Separator)) {
		dir = path
		if dir == "" {
			dir = "."
		}
		prefix = ""
	} else {
		dir = filepath.Dir(path)
		prefix = filepath.Base(path)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var matches []string
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		full := filepath.Join(dir, e.Name())
		if e.IsDir() {
			full += string(filepath.Separator)
		}
		matches = append(matches, full)
	}
	return matches
}

// expandHome replaces a leading "~" with the current user's home directory.
func (p *PathInput) expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	// Use string concatenation so that a leading "/" in path[1:] is preserved
	// correctly (e.g. "~/foo" → home + "/foo", not filepath.Join which would
	// treat "/foo" as an absolute path).
	return home + path[1:]
}

// longestCommonPrefix returns the longest string that is a prefix of every element
// in strs.
func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	for _, s := range strs[1:] {
		for !strings.HasPrefix(s, prefix) {
			if len(prefix) == 0 {
				return ""
			}
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}
