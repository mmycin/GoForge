// Package confirm provides a reusable yes/no confirmation dialog for GoForge TUI screens.
package confirm

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mmycin/GoForge/internal/tui"
)

// Model is a minimal yes/no confirm dialog.
// When destructive=true the default focus is "No" and the Yes button is red.
type Model struct {
	title       string
	body        string
	destructive bool
	focused     bool // true = Yes focused, false = No focused
	done        bool
}

// New creates a new confirm Model.
// When destructive=true the default focus is "No, keep it".
// When destructive=false the default focus is "Yes, proceed".
func New(title, body string, destructive bool) Model {
	return Model{
		title:       title,
		body:        body,
		destructive: destructive,
		focused:     !destructive, // Yes focused only for non-destructive
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
// On confirm or cancel, sets done and returns tea.Quit so the program exits.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "tab", "left", "right", "h", "l":
			m.focused = !m.focused
			return m, nil
		case "enter", " ":
			m.done = true
			return m, tea.Quit
		case "esc", "q", "ctrl+c":
			m.focused = false
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// Confirmed reports whether the user selected Yes.
// Safe to call after tea.Program.Run() returns.
func (m Model) Confirmed() bool { return m.done && m.focused }

// View implements tea.Model.
func (m Model) View() string {
	if m.done {
		return ""
	}

	headerIcon := tui.IconDanger
	if !m.destructive {
		headerIcon = tui.IconInfo
	}
	header := tui.Header(headerIcon + "  " + m.title)

	bodyStyle := lipgloss.NewStyle().Padding(1, 2)
	content := bodyStyle.Render(m.body)

	yesStyle := lipgloss.NewStyle().Padding(0, 2).Bold(true)
	noStyle := lipgloss.NewStyle().Padding(0, 2).Bold(true)

	if m.focused { // Yes focused
		if m.destructive {
			yesStyle = yesStyle.Background(tui.ErrorColor).Foreground(lipgloss.Color("#FFFFFF"))
		} else {
			yesStyle = yesStyle.Background(tui.SuccessColor).Foreground(lipgloss.Color("#000000"))
		}
		noStyle = noStyle.Foreground(tui.MutedColor)
	} else { // No focused
		noStyle = noStyle.Background(tui.PrimaryColor).Foreground(lipgloss.Color("#FFFFFF"))
		yesStyle = yesStyle.Foreground(tui.MutedColor)
	}

	buttons := lipgloss.NewStyle().Padding(1, 2).Render(
		noStyle.Render("  No, keep it  ") + "   " + yesStyle.Render("  Yes, proceed  "),
	)

	footer := tui.Footer("tab", "toggle", "enter", "confirm", "esc", "cancel")
	return header + "\n\n" + content + "\n" + buttons + "\n\n" + footer
}
