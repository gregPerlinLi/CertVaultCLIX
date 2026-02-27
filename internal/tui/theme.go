package tui

// This file re-exports styles from the styles subpackage.
// Use the styles subpackage directly in components and views to avoid import cycles.

import s "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"

// Re-export all styles for use within the tui package itself.
var (
TitleStyle           = s.TitleStyle
SubtitleStyle        = s.SubtitleStyle
SelectedStyle        = s.SelectedStyle
NormalStyle          = s.NormalStyle
MutedStyle           = s.MutedStyle
SuccessStyle         = s.SuccessStyle
WarningStyle         = s.WarningStyle
DangerStyle          = s.DangerStyle
BorderStyle          = s.BorderStyle
FocusBorderStyle     = s.FocusBorderStyle
StatusBarStyle       = s.StatusBarStyle
StatusBarHighlight   = s.StatusBarHighlight
HelpStyle            = s.HelpStyle
KeyStyle             = s.KeyStyle
PaginationStyle      = s.PaginationStyle
SidebarStyle         = s.SidebarStyle
SidebarItemStyle     = s.SidebarItemStyle
SidebarSelectedStyle = s.SidebarSelectedStyle
SidebarHeaderStyle   = s.SidebarHeaderStyle
DialogStyle          = s.DialogStyle
ButtonStyle          = s.ButtonStyle
ButtonInactiveStyle  = s.ButtonInactiveStyle
InputStyle           = s.InputStyle
InputFocusStyle      = s.InputFocusStyle
TableHeaderStyle     = s.TableHeaderStyle
TableRowStyle        = s.TableRowStyle
TableSelectedRowStyle = s.TableSelectedRowStyle
ToastSuccessStyle    = s.ToastSuccessStyle
ToastErrorStyle      = s.ToastErrorStyle
ToastInfoStyle       = s.ToastInfoStyle
)

// ExpiryStyle returns a color-coded style based on days remaining.
var ExpiryStyle = s.ExpiryStyle

// RoleStyle returns a color-coded style based on user role.
var RoleStyle = s.RoleStyle
