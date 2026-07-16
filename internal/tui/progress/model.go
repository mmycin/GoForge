// Package progress provides a multi-step progress tracker for long-running operations.
package progress

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mmycin/GoForge/internal/tui"
)

// StepState represents the lifecycle of a single step.
type StepState int

const (
	StatePending StepState = iota
	StateRunning
	StateDone
	StateFailed
)

// Step is a named unit of work.
type Step struct {
	Label string
	State StepState
	Err   error
}

// Model tracks multiple named steps and renders them with a progress bar.
type Model struct {
	title   string
	steps   []Step
	bar     progress.Model
	spinner spinner.Model
	done    bool
}

// New creates a progress Model for the given step labels.
func New(title string, labels []string) Model {
	steps := make([]Step, len(labels))
	for i, l := range labels {
		steps[i] = Step{Label: l, State: StatePending}
	}
	bar := progress.New(
		progress.WithGradient(string(tui.PrimaryColor), string(tui.SecondaryColor)),
		progress.WithoutPercentage(),
	)
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = tui.InfoStyle
	return Model{title: title, steps: steps, bar: bar, spinner: sp}
}

// Advance marks the step at index as Running.
func (m *Model) Advance(idx int) {
	if idx >= 0 && idx < len(m.steps) {
		m.steps[idx].State = StateRunning
	}
}

// Complete marks the step at index as Done.
func (m *Model) Complete(idx int) {
	if idx >= 0 && idx < len(m.steps) {
		m.steps[idx].State = StateDone
	}
}

// Fail marks the step at index as Failed.
func (m *Model) Fail(idx int, err error) {
	if idx >= 0 && idx < len(m.steps) {
		m.steps[idx].State = StateFailed
		m.steps[idx].Err = err
	}
}

// SetDone marks all work as finished.
func (m *Model) SetDone() { m.done = true }

// IndexOf finds the first step whose Label contains substr (case-sensitive).
func (m *Model) IndexOf(label string) int {
	for i, s := range m.steps {
		if strings.Contains(s.Label, label) {
			return i
		}
	}
	return -1
}

// Percent returns the fraction (0..1) of completed/failed steps.
func (m *Model) Percent() float64 {
	if len(m.steps) == 0 {
		return 0
	}
	done := 0
	for _, s := range m.steps {
		if s.State == StateDone || s.State == StateFailed {
			done++
		}
	}
	return float64(done) / float64(len(m.steps))
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update implements tea.Model — must return (tea.Model, tea.Cmd).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case progress.FrameMsg:
		pm, cmd := m.bar.Update(msg)
		m.bar = pm.(progress.Model)
		cmds = append(cmds, cmd)
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the step list and progress bar.
func (m Model) View() string {
	// Pass only the title — Header() already adds the ⚡ GoForge brand on the left.
	header := tui.Header(m.title)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n  %s\n\n", tui.HeadingStyle.Render(m.title)))

	for _, step := range m.steps {
		icon := tui.StepIcon("pending")
		label := tui.MutedStyle.Render(step.Label)
		suffix := ""

		switch step.State {
		case StateRunning:
			icon = m.spinner.View()
			label = lipgloss.NewStyle().Foreground(tui.SecondaryColor).Render(step.Label)
		case StateDone:
			icon = tui.StepIcon("done")
			label = step.Label
		case StateFailed:
			icon = tui.StepIcon("failed")
			label = tui.ErrorStyle.Render(step.Label)
			if step.Err != nil {
				suffix = "  " + tui.MutedStyle.Render(step.Err.Error())
			}
		}
		sb.WriteString(fmt.Sprintf("  %s  %s%s\n", icon, label, suffix))
	}

	sb.WriteString("\n  " + m.bar.ViewAs(m.Percent()) + "\n")

	if m.done {
		sb.WriteString("\n  " + tui.SuccessStyle.Render(tui.IconSuccess+"  All steps completed") + "\n")
	}

	footer := tui.Footer("please wait", "this may take a moment")
	if m.done {
		footer = tui.Footer("enter", "dismiss", "q", "quit")
	}

	return header + "\n" + sb.String() + "\n" + footer
}
