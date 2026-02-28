package components

import (
"fmt"
"strings"

tea "github.com/charmbracelet/bubbletea"
"github.com/charmbracelet/lipgloss"
st "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
)

// Column defines a table column.
type Column struct {
Title string
Width int
}

// Row is a table row (slice of strings).
type Row []string

// Table is an interactive, keyboard-navigable table component.
type Table struct {
Columns []Column
Rows    []Row
cursor  int
offset  int
height  int
width   int
focused bool
}

// NewTable creates a new table.
func NewTable(cols []Column, height int) Table {
return Table{
Columns: cols,
height:  height,
focused: true,
}
}

// SetRows replaces the table rows.
func (t *Table) SetRows(rows []Row) {
t.Rows = rows
if t.cursor >= len(rows) && len(rows) > 0 {
t.cursor = len(rows) - 1
} else if len(rows) == 0 {
t.cursor = 0
}
t.offset = 0
}

// SetSize sets the table dimensions.
func (t *Table) SetSize(width, height int) {
t.width = width
t.height = height
}

// SetFocused sets the focus state.
func (t *Table) SetFocused(f bool) {
t.focused = f
}

// SelectedRow returns the currently selected row.
func (t *Table) SelectedRow() (Row, bool) {
if len(t.Rows) == 0 || t.cursor < 0 || t.cursor >= len(t.Rows) {
return nil, false
}
return t.Rows[t.cursor], true
}

// SelectedIndex returns the cursor index.
func (t *Table) SelectedIndex() int {
return t.cursor
}

// Update handles keyboard and mouse events.
func (t *Table) Update(msg tea.Msg) tea.Cmd {
switch msg := msg.(type) {
case tea.KeyMsg:
switch msg.String() {
case "up", "k":
if t.cursor > 0 {
t.cursor--
if t.cursor < t.offset {
t.offset--
}
}
case "down", "j":
if t.cursor < len(t.Rows)-1 {
t.cursor++
if t.cursor >= t.offset+t.visibleRows() {
t.offset++
}
}
case "ctrl+u":
// Half-page scroll up (vim-style: move cursor AND offset together)
step := t.visibleRows() / 2
if step < 1 {
step = 1
}
t.cursor -= step
t.offset -= step
if t.cursor < 0 {
t.cursor = 0
}
if t.offset < 0 {
t.offset = 0
}
if t.cursor < t.offset {
t.cursor = t.offset
}
case "ctrl+d":
// Half-page scroll down (vim-style: move cursor AND offset together)
step := t.visibleRows() / 2
if step < 1 {
step = 1
}
t.cursor += step
t.offset += step
if len(t.Rows) > 0 {
if t.cursor >= len(t.Rows) {
t.cursor = len(t.Rows) - 1
}
maxOffset := len(t.Rows) - t.visibleRows()
if maxOffset < 0 {
maxOffset = 0
}
if t.offset > maxOffset {
t.offset = maxOffset
}
if t.cursor < t.offset {
t.cursor = t.offset
}
}
}
case tea.MouseMsg:
switch msg.Button {
case tea.MouseButtonWheelUp:
if t.cursor > 0 {
t.cursor--
if t.cursor < t.offset {
t.offset--
}
}
case tea.MouseButtonWheelDown:
if t.cursor < len(t.Rows)-1 {
t.cursor++
if t.cursor >= t.offset+t.visibleRows() {
t.offset++
}
}
}
}
return nil
}

func (t *Table) visibleRows() int {
if t.height > 2 {
return t.height - 2 // subtract header + separator
}
return t.height
}

// View renders the table.
func (t *Table) View() string {
var sb strings.Builder

// Header
header := t.renderRow(t.headerRow(), true, false)
sb.WriteString(header)
sb.WriteString("\n")

// Separator
sep := strings.Repeat("─", t.totalWidth())
sb.WriteString(st.MutedStyle.Render(sep))
sb.WriteString("\n")

// Body
visible := t.visibleRows()
end := t.offset + visible
if end > len(t.Rows) {
end = len(t.Rows)
}
for i := t.offset; i < end; i++ {
selected := t.focused && i == t.cursor
sb.WriteString(t.renderRow(t.Rows[i], false, selected))
sb.WriteString("\n")
}

// Empty state
if len(t.Rows) == 0 {
empty := st.MutedStyle.Render("  No Data")
sb.WriteString(empty)
sb.WriteString("\n")
}

// Pagination indicator
if len(t.Rows) > 0 {
info := fmt.Sprintf("  %d/%d", t.cursor+1, len(t.Rows))
sb.WriteString(st.PaginationStyle.Render(info))
}

return sb.String()
}

func (t *Table) headerRow() Row {
r := make(Row, len(t.Columns))
for i, c := range t.Columns {
r[i] = c.Title
}
return r
}

// renderRow renders a single table row. Cells may contain pre-applied ANSI color codes;
// lipgloss.Width is used for visual-width-aware padding and truncation so that styled
// cells (role colors, expiry colors) remain aligned with plain-text header cells.
func (t *Table) renderRow(row Row, isHeader, isSelected bool) string {
var cells []string
for i, col := range t.Columns {
var cell string
if i < len(row) {
cell = row[i]
}

visWidth := lipgloss.Width(cell)

// Detect plain text by comparing visual width to rune count.
// For plain ASCII/Unicode text these are equal; ANSI codes add bytes but no visual width.
runeCount := len([]rune(cell))
isPlainText := visWidth == runeCount
if isPlainText && visWidth > col.Width-1 {
runes := []rune(cell)
if len(runes) > col.Width-4 {
cell = string(runes[:col.Width-4]) + "..."
}
visWidth = lipgloss.Width(cell)
}

// Pad to column width using visual width (correct for CJK and ANSI-coded cells).
if pad := col.Width - visWidth; pad > 0 {
cell = cell + strings.Repeat(" ", pad)
}

var style lipgloss.Style
switch {
case isHeader:
style = st.TableHeaderStyle
case isSelected:
style = st.TableSelectedRowStyle
default:
style = st.TableRowStyle
}
cells = append(cells, style.Render(cell))
}
return strings.Join(cells, " ")
}

func (t *Table) totalWidth() int {
total := 0
for _, c := range t.Columns {
total += c.Width + 1
}
return total
}
