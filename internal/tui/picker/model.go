// Package picker provides a list-based item picker for choosing from a dynamic
// set of strings (e.g. service names discovered from the file system).
package picker

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mmycin/GoForge/internal/tui"
)

// item implements list.Item.
type item struct {
	value string
	desc  string
}

func (i item) FilterValue() string { return i.value }
func (i item) Title() string       { return i.value }
func (i item) Description() string { return i.desc }

// itemDelegate renders each list item with GoForge styling.
type itemDelegate struct{}

func (d itemDelegate) Height() int                              { return 1 }
func (d itemDelegate) Spacing() int                             { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	cursor := "   "
	name := i.value
	desc := ""
	if i.desc != "" {
		desc = tui.MutedStyle.Render("  " + i.desc)
	}
	if index == m.Index() {
		cursor = tui.InfoStyle.Render(" " + tui.IconSelect + " ")
		name = lipgloss.NewStyle().Foreground(tui.SecondaryColor).Bold(true).Render(i.value)
	}
	fmt.Fprintf(w, "%s%s%s", cursor, name, desc)
}

// Model is the picker component.
// After tea.Program.Run() returns, call Selected() to get the chosen value
// (empty string means cancelled).
type Model struct {
	list     list.Model
	title    string
	selected string // set before tea.Quit so it survives in the final model
	done     bool
}

// New creates a picker Model.
func New(title string, values []string, descs []string) Model {
	items := make([]list.Item, len(values))
	for i, v := range values {
		desc := ""
		if i < len(descs) {
			desc = descs[i]
		}
		items[i] = item{value: v, desc: desc}
	}

	l := list.New(items, itemDelegate{}, 60, min(len(values)+4, 20))
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(tui.SecondaryColor)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(tui.PrimaryColor)

	return Model{list: l, title: title}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
// On Enter: records the selection and calls tea.Quit — the final model is
// readable immediately after tea.Program.Run() returns.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.list.FilterState() != list.Filtering {
				if sel, ok := m.list.SelectedItem().(item); ok {
					m.selected = sel.value
					m.done = true
					return m, tea.Quit
				}
			}
		case "esc", "q", "ctrl+c":
			if m.list.FilterState() == list.Filtering {
				// fall through — let list handle clearing the filter
			} else {
				m.done = true
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width - 4)
		m.list.SetHeight(msg.Height - 6)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	if m.done {
		return ""
	}

	header := tui.Header(m.title)
	count := fmt.Sprintf("\n  %s\n\n", tui.MutedStyle.Render(
		fmt.Sprintf("%d items", len(m.list.Items())),
	))

	listView := strings.Split(m.list.View(), "\n")
	var indented []string
	for _, l := range listView {
		indented = append(indented, "  "+l)
	}

	footer := tui.Footer("↑↓", "navigate", "enter", "select", "/", "filter", "esc", "cancel")
	return header + count + strings.Join(indented, "\n") + "\n\n" + footer
}

// Selected returns the value chosen by the user, or "" if cancelled.
// Safe to call after tea.Program.Run() returns.
func (m Model) Selected() string { return m.selected }

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
