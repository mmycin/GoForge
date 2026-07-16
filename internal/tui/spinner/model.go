// Package spinner wraps bubbles/spinner with a label and GoForge theme styling.
package spinner

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmycin/GoForge/internal/tui"
)

// Model is a labelled spinner widget.
type Model struct {
	sp    spinner.Model
	Label string
}

// New creates a new spinner Model with the given label.
func New(label string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = tui.InfoStyle
	return Model{sp: s, Label: label}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return m.sp.Tick }

// Update implements tea.Model — must return (tea.Model, tea.Cmd).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.sp, cmd = m.sp.Update(msg)
	return m, cmd
}

// View returns the rendered spinner + label.
func (m Model) View() string {
	return m.sp.View() + "  " + tui.MutedStyle.Render(m.Label)
}
