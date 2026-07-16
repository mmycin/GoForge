package dashboard

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mmycin/GoForge/internal/tui"
)

// ── Data ──────────────────────────────────────────────────────────────────────

type section struct {
	label    string
	commands []CommandItem
}

var sections = []section{
	{
		label: "Project",
		commands: []CommandItem{
			{Name: "new", Desc: "Create a new GoForge project", HasWizard: true},
			{Name: "app serve", Desc: "Start the development server"},
		},
	},
	{
		label: "Code Generation",
		commands: []CommandItem{
			{Name: "gen:service", Desc: "Scaffold a new service layer", HasWizard: true},
			{Name: "gen:event", Desc: "Generate event structs & listeners", HasWizard: true},
			{Name: "gen:command", Desc: "Create a custom console command", HasWizard: true},
			{Name: "gen:config", Desc: "Add a new config section", HasWizard: true},
			{Name: "gen:proto", Desc: "Compile proto files & generate gRPC stubs", HasWizard: true},
		},
	},
	{
		label: "Database",
		commands: []CommandItem{
			{Name: "migrate", Desc: "Apply pending migrations"},
			{Name: "gen:migration", Desc: "Create a new migration file", HasWizard: true},
			{Name: "rem:migration", Desc: "Remove the latest migration"},
			{Name: "gen:sqlc", Desc: "Generate type-safe SQL query code"},
			{Name: "rem:sqlc", Desc: "Remove generated SQLC integration"},
			{Name: "loader", Desc: "Inspect GORM schema output"},
		},
	},
	{
		label: "Maintenance",
		commands: []CommandItem{
			{Name: "gen:key", Desc: "Generate & save a secure APP_KEY"},
			{Name: "rem:key", Desc: "Clear the APP_KEY from .env"},
			{Name: "version", Desc: "Show version and build info"},
			{Name: "readme", Desc: "View the recommended workflow guide"},
		},
	},
}

// flat is the ordered list of selectable commands across all sections.
// Built once at init time — sections are rendered separately as decorations.
var flat []CommandItem

func init() {
	for _, s := range sections {
		flat = append(flat, s.commands...)
	}
}

// ── Model ─────────────────────────────────────────────────────────────────────
// Model is the root interactive command browser.
// It owns its own keyboard handling — no bubbles/list involved.
type Model struct {
	cursor   int  // index into flat[]
	quitting bool
	selected CommandItem // set on Enter before tea.Quit
	// filter support
	filtering bool
	filter    string
	filtered  []CommandItem // nil = show all
}

// New constructs the dashboard Model.
func New() Model {
	return Model{}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// visibleList returns the current effective command list.
func (m Model) visibleList() []CommandItem {
	if m.filter == "" {
		return flat
	}
	return m.filtered
}

// clampCursor keeps cursor inside the visible list.
func (m *Model) clampCursor() {
	list := m.visibleList()
	if len(list) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(list) {
		m.cursor = len(list) - 1
	}
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// ── filter mode ───────────────────────────────────────────────────────
		if m.filtering {
			switch msg.String() {
			case "esc", "ctrl+c":
				m.filtering = false
				m.filter = ""
				m.filtered = nil
				m.cursor = 0
			case "enter":
				m.filtering = false
			case "backspace":
				if len(m.filter) > 0 {
					m.filter = m.filter[:len(m.filter)-1]
					m.rebuildFilter()
				}
			default:
				if len(msg.Runes) > 0 {
					m.filter += string(msg.Runes)
					m.rebuildFilter()
				}
			}
			return m, nil
		}

		// ── normal mode ───────────────────────────────────────────────────────
		switch msg.String() {
		case "up", "k":
			m.cursor--
			m.clampCursor()
		case "down", "j":
			m.cursor++
			m.clampCursor()
		case "/":
			m.filtering = true
			m.filter = ""
			m.filtered = nil
			m.cursor = 0
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter", " ":
			list := m.visibleList()
			if len(list) > 0 && m.cursor < len(list) {
				m.selected = list[m.cursor]
				m.quitting = true
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		// nothing to resize — pure text rendering
	}
	return m, nil
}

func (m *Model) rebuildFilter() {
	q := strings.ToLower(m.filter)
	m.filtered = nil
	for _, cmd := range flat {
		if strings.Contains(strings.ToLower(cmd.Name), q) ||
			strings.Contains(strings.ToLower(cmd.Desc), q) {
			m.filtered = append(m.filtered, cmd)
		}
	}
	m.cursor = 0
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	header := tui.Header("Command Center")

	var sb strings.Builder
	sb.WriteString("\n")

	list := m.visibleList()

	if m.filter != "" {
		// flat filtered view — no section headers
		sb.WriteString(fmt.Sprintf("  %s\n\n",
			tui.MutedStyle.Render("Filter: ")+tui.InfoStyle.Render(m.filter),
		))
		if len(list) == 0 {
			sb.WriteString("  " + tui.MutedStyle.Render("No commands match.") + "\n")
		}
		for i, cmd := range list {
			sb.WriteString(renderRow(cmd, i == m.cursor))
		}
	} else {
		// sectioned view
		idx := 0
		for _, sec := range sections {
			sb.WriteString("  " + tui.HeadingStyle.Render(sec.label) + "\n")
			for _, cmd := range sec.commands {
				sb.WriteString(renderRow(cmd, idx == m.cursor))
				idx++
			}
			sb.WriteString("\n")
		}
	}

	// filter input when active
	if m.filtering {
		sb.WriteString("  " + tui.InfoStyle.Render("/") + " " +
			tui.InfoStyle.Render(m.filter+"█") + "\n")
	}

	footer := tui.Footer("↑↓", "navigate", "enter", "run", "/", "filter", "q", "quit")
	return header + "\n" + sb.String() + footer
}

func renderRow(cmd CommandItem, selected bool) string {
	cursor := "     "
	nameStyle := lipgloss.NewStyle().Width(20)
	descStyle := tui.MutedStyle.Copy()

	if selected {
		cursor = "  " + tui.InfoStyle.Render(tui.IconSelect) + "  "
		nameStyle = nameStyle.Foreground(tui.SecondaryColor).Bold(true)
		descStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#D1D5DB"))
	}

	return fmt.Sprintf("%s%s  %s\n",
		cursor,
		nameStyle.Render(cmd.Name),
		descStyle.Render(cmd.Desc),
	)
}

// Selected returns the command chosen by the user (zero value if quit without selecting).
func (m Model) Selected() CommandItem { return m.selected }

// Items returns the full flat command list.
func (m Model) Items() []CommandItem { return flat }

// Ensure Model satisfies tea.Model at compile time.
var _ tea.Model = Model{}
