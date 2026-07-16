// Package viewport wraps bubbles/viewport to provide a live-output scrollable
// log panel for commands that stream subprocess output.
package viewport

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmycin/GoForge/internal/tui"
)

// AppendLineMsg adds a single line to the viewport.
type AppendLineMsg struct{ Line string }

// DoneMsg signals the subprocess has finished.
type DoneMsg struct{ Err error }

// Model wraps bubbles/viewport for live-output display.
type Model struct {
	vp        viewport.Model
	title     string
	lines     []string
	streaming bool
	err       error
	width     int
	height    int
}

// New creates a viewport Model.
func New(title string, width, height int) Model {
	vp := viewport.New(width, height-6) // leave room for header/footer
	vp.Style = tui.ContentStyle
	return Model{
		vp:        vp,
		title:     title,
		streaming: true,
		width:     width,
		height:    height,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model — returns (tea.Model, tea.Cmd) to satisfy the interface.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case AppendLineMsg:
		styled := tui.InfoStyle.Render(tui.IconInfo+"  ") + msg.Line
		m.lines = append(m.lines, styled)
		m.vp.SetContent(strings.Join(m.lines, "\n"))
		if m.streaming {
			m.vp.GotoBottom()
		}

	case DoneMsg:
		m.streaming = false
		m.err = msg.Err
		if msg.Err != nil {
			errLine := tui.ErrorStyle.Render(tui.IconError+"  Command failed: "+msg.Err.Error())
			m.lines = append(m.lines, errLine)
		} else {
			m.lines = append(m.lines, tui.SuccessStyle.Render(tui.IconSuccess+"  Completed successfully"))
		}
		m.vp.SetContent(strings.Join(m.lines, "\n"))
		m.vp.GotoBottom()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.vp.Width = msg.Width - 4
		m.vp.Height = msg.Height - 6

	default:
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		cmds = append(cmds, cmd)

		if km, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(km, key.NewBinding(key.WithKeys("up", "k"))):
				m.vp.LineUp(1)
			case key.Matches(km, key.NewBinding(key.WithKeys("down", "j"))):
				m.vp.LineDown(1)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// UpdateModel is a type-safe convenience wrapper for callers that need the
// concrete Model back without writing a type assertion every time.
func UpdateModel(m Model, msg tea.Msg) (Model, tea.Cmd) {
	updated, cmd := m.Update(msg)
	if nm, ok := updated.(Model); ok {
		return nm, cmd
	}
	return m, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	header := tui.Header(m.title)
	content := m.vp.View()

	scrollHint := ""
	if len(m.lines) > 0 && !m.streaming && m.vp.ScrollPercent() < 1.0 {
		scrollHint = "\n" + tui.MutedStyle.Render("  ↓ more output below")
	}

	var footer string
	if m.streaming {
		footer = tui.Footer("↑↓", "scroll", "please wait", "streaming output")
	} else {
		footer = tui.Footer("↑↓", "scroll", "enter", "dismiss", "q", "quit")
	}

	return header + "\n\n" + content + scrollHint + "\n\n" + footer
}

// IsStreaming reports whether the subprocess is still running.
func (m Model) IsStreaming() bool { return m.streaming }

// HasError reports whether the subprocess ended with an error.
func (m Model) HasError() bool { return m.err != nil }
