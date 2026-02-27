package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	st "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
)

// ToastType represents the type of toast notification.
type ToastType int

const (
	ToastSuccess ToastType = iota
	ToastError
	ToastInfo
)

// ClearToastMsg signals that the toast should be cleared.
type ClearToastMsg struct{}

// Toast is a non-blocking notification component.
type Toast struct {
	message  string
	toastType ToastType
	visible  bool
}

// Show displays a toast notification.
func (t *Toast) Show(msg string, tp ToastType) tea.Cmd {
	t.message = msg
	t.toastType = tp
	t.visible = true
	return tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
		return ClearToastMsg{}
	})
}

// Hide hides the toast.
func (t *Toast) Hide() {
	t.visible = false
}

// IsVisible returns whether the toast is visible.
func (t *Toast) IsVisible() bool {
	return t.visible
}

// View renders the toast.
func (t *Toast) View() string {
	if !t.visible {
		return ""
	}
	switch t.toastType {
	case ToastSuccess:
		return st.ToastSuccessStyle.Render("✓ " + t.message)
	case ToastError:
		return st.ToastErrorStyle.Render("✗ " + t.message)
	default:
		return st.ToastInfoStyle.Render("ℹ " + t.message)
	}
}
