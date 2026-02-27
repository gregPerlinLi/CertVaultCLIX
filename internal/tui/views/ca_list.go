package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
	tui "github.com/gregPerlinLi/CertVaultCLIX/internal/tui/styles"
	"github.com/gregPerlinLi/CertVaultCLIX/internal/tui/components"
)

// CAListLoadedMsg carries CA list data.
type CAListLoadedMsg struct {
	CAs   []api.CACert
	Total int64
	Err   error
}

// CAList is the CA certificate list view.
type CAList struct {
	client  *api.Client
	table   components.Table
	cas     []api.CACert
	total   int64
	page    int
	spinner components.Spinner
	toast   components.Toast
	err     string
	width   int
	height  int
}

// NewCAList creates a new CA list view.
func NewCAList(client *api.Client) CAList {
	cols := []components.Column{
		{Title: "CN", Width: 30},
		{Title: "Algorithm", Width: 12},
		{Title: "Expires", Width: 20},
		{Title: "Available", Width: 10},
		{Title: "Comment", Width: 20},
	}
	return CAList{
		client:  client,
		table:   components.NewTable(cols, 15),
		page:    1,
		spinner: components.NewSpinner(),
	}
}

// SetSize updates dimensions.
func (c *CAList) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.table.SetSize(width, height-5)
}

// Init loads CAs.
func (c *CAList) Init() tea.Cmd {
	cmd := c.spinner.Start("Loading CA certificates...")
	return tea.Batch(cmd, c.load())
}

func (c *CAList) load() tea.Cmd {
	return func() tea.Msg {
		cas, err := c.client.ListUserCAs(context.Background(), c.page, 20)
		if err != nil {
			return CAListLoadedMsg{Err: err}
		}
		return CAListLoadedMsg{CAs: cas.List, Total: cas.Total}
	}
}

// SelectedCA returns the currently selected CA or nil.
func (c *CAList) SelectedCA() *api.CACert {
	row, ok := c.table.SelectedRow()
	if !ok || len(c.cas) == 0 {
		return nil
	}
	_ = row
	idx := c.table.SelectedIndex()
	if idx >= 0 && idx < len(c.cas) {
		return &c.cas[idx]
	}
	return nil
}

// Update handles messages.
func (c *CAList) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case CAListLoadedMsg:
		c.spinner.Stop()
		if msg.Err != nil {
			c.err = msg.Err.Error()
			return nil
		}
		c.cas = msg.CAs
		c.total = msg.Total
		c.table.SetRows(c.buildRows())
		return nil

	case tea.KeyMsg:
		switch msg.String() {
		case "r", "f5":
			cmd := c.spinner.Start("Refreshing...")
			return tea.Batch(cmd, c.load())
		case "pgup":
			if c.page > 1 {
				c.page--
				return c.load()
			}
		case "pgdown":
			if int64(c.page*20) < c.total {
				c.page++
				return c.load()
			}
		default:
			return c.table.Update(msg)
		}
	}
	return c.spinner.Update(msg)
}

func (c *CAList) buildRows() []components.Row {
	rows := make([]components.Row, len(c.cas))
	for i, ca := range c.cas {
		avail := "✓"
		if !ca.Available {
			avail = "✗"
		}
		rows[i] = components.Row{
			ca.CN,
			ca.Algorithm,
			formatExpiry(ca.NotAfter),
			avail,
			ca.Comment,
		}
	}
	return rows
}

// View renders the CA list.
func (c *CAList) View() string {
	var sb strings.Builder
	sb.WriteString(tui.TitleStyle.Render("🔐 CA Certificates"))
	sb.WriteString("\n\n")

	if c.spinner.IsActive() {
		sb.WriteString(c.spinner.View())
		return sb.String()
	}
	if c.err != "" {
		sb.WriteString(tui.DangerStyle.Render("Error: " + c.err))
		sb.WriteString("\n")
		sb.WriteString(tui.HelpStyle.Render("Press r to retry"))
		return sb.String()
	}

	total := fmt.Sprintf("Total: %d | Page: %d", c.total, c.page)
	sb.WriteString(tui.MutedStyle.Render(total))
	sb.WriteString("\n")
	sb.WriteString(c.table.View())
	sb.WriteString("\n")
	sb.WriteString(tui.HelpStyle.Render("enter: details • r: refresh • PgUp/PgDn: page"))

	if c.toast.IsVisible() {
		sb.WriteString("\n" + c.toast.View())
	}
	return sb.String()
}

func formatExpiry(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Format("2006-01-02")
}
