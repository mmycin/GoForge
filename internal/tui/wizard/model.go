// Package wizard wraps charmbracelet/huh forms with GoForge chrome rendering.
package wizard

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/mmycin/GoForge/internal/tui"
)

// Model wraps a huh.Form with GoForge chrome.
type Model struct {
	title      string
	totalSteps int
	form       *huh.Form
	done       bool
	cancelled  bool
}

// New creates a wizard Model.
func New(title string, totalSteps int, form *huh.Form) Model {
	return Model{title: title, totalSteps: totalSteps, form: form}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return m.form.Init() }

// Update implements tea.Model — must return (tea.Model, tea.Cmd).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done || m.cancelled {
		return m, nil
	}
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	if m.form.State == huh.StateCompleted {
		m.done = true
		return m, tea.Quit
	}
	if m.form.State == huh.StateAborted {
		m.cancelled = true
		return m, tea.Quit
	}
	return m, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	if m.done || m.cancelled {
		return ""
	}
	header := tui.Header(m.title, fmt.Sprintf("Step ? of %d", m.totalSteps))
	footer := tui.Footer("enter", "continue", "esc", "back")
	return header + "\n\n" + m.form.View() + "\n" + footer
}

// IsDone reports whether the form was submitted.
func (m Model) IsDone() bool { return m.done }

// IsCancelled reports whether the form was cancelled.
func (m Model) IsCancelled() bool { return m.cancelled }
