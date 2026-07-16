// Package result provides a post-operation result panel that lists generated
// files and an optional "next steps" hint.
package result

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mmycin/GoForge/internal/tui"
	"github.com/mmycin/GoForge/internal/usecase"
)

// Model renders a dismissible result panel.
type Model struct {
	title    string
	files    []usecase.GeneratedFile
	nextStep string
}

// New creates a new result Model.
func New(title string, files []usecase.GeneratedFile, nextStep string) Model {
	return Model{title: title, files: files, nextStep: nextStep}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
// Any dismiss key quits the program immediately via tea.Quit.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "enter", "q", "Q", " ", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	// Pass plain title — Header() adds the ⚡ GoForge brand.
	header := tui.Header(tui.IconSuccess + "  " + m.title)

	var sb strings.Builder
	sb.WriteString("\n")

	if len(m.files) > 0 {
		createdLabel := lipgloss.NewStyle().Foreground(tui.MutedColor).Render("  Generated files:")
		updatedLabel := lipgloss.NewStyle().Foreground(tui.MutedColor).Render("  Updated:")

		var createdList, updatedList []string
		for _, f := range m.files {
			line := "    " + tui.InfoStyle.Render(tui.IconInfo) + "  " + tui.CodeStyle.Render(f.Path)
			if f.Updated {
				updatedList = append(updatedList, line)
			} else {
				createdList = append(createdList, line)
			}
		}
		if len(createdList) > 0 {
			sb.WriteString(createdLabel + "\n")
			sb.WriteString(strings.Join(createdList, "\n") + "\n\n")
		}
		if len(updatedList) > 0 {
			sb.WriteString(updatedLabel + "\n")
			sb.WriteString(strings.Join(updatedList, "\n") + "\n\n")
		}
	}

	if m.nextStep != "" {
		sb.WriteString("  " + tui.MutedStyle.Render("Next:") + "  " + tui.CodeStyle.Render(m.nextStep) + "\n\n")
	}

	footer := tui.Footer("enter", "dismiss", "q", "quit")
	return header + "\n" + sb.String() + footer
}
