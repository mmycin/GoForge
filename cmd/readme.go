package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mmycin/GoForge/internal/tui"
	"github.com/spf13/cobra"
)

func newReadmeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "readme",
		Short: "View the recommended workflow guide",
		Long:  `Display a scrollable workflow guide for getting started with GoForge.`,
		Run: func(cmd *cobra.Command, args []string) {
			p := tea.NewProgram(newReadmeModel(), tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				ErrorLog("readme: %v", err)
			}
		},
	}
}

// ── Readme model ─────────────────────────────────────────────────────────────

type readmeModel struct {
	vp     viewport.Model
	ready  bool
}

func newReadmeModel() readmeModel { return readmeModel{} }

func (m readmeModel) Init() tea.Cmd { return nil }

func (m readmeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.vp = viewport.New(msg.Width-4, msg.Height-6)
			m.vp.SetContent(readmeContent())
			m.ready = true
		} else {
			m.vp.Width = msg.Width - 4
			m.vp.Height = msg.Height - 6
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m readmeModel) View() string {
	if !m.ready {
		return tui.MutedStyle.Render("\n  Loading…")
	}
	header := tui.Header("Recommended Workflow")
	footer := tui.Footer("↑↓", "scroll", "q", "quit")
	return header + "\n\n" + m.vp.View() + "\n\n" + footer
}

// readmeContent returns the styled workflow document rendered as a string.
func readmeContent() string {
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.PrimaryColor)
	cmdStyle := tui.CodeStyle.Copy().Width(32)
	descStyle := tui.MutedStyle.Copy()

	row := func(n int, cmd, desc string) string {
		num := lipgloss.NewStyle().Foreground(tui.MutedColor).Width(4).Render(fmt.Sprintf("%d", n))
		return fmt.Sprintf("  %s%s  %s", num, cmdStyle.Render(cmd), descStyle.Render(desc))
	}

	var sb strings.Builder

	sb.WriteString(sectionStyle.Render("── Starting a new project") + "\n\n")
	sb.WriteString(row(1, "goforge new", "Create a new project") + "\n")
	sb.WriteString(row(2, "goforge gen:key", "Generate APP_KEY") + "\n")
	sb.WriteString(row(3, "goforge gen:migration init", "Initialise migration history") + "\n")
	sb.WriteString(row(4, "goforge migrate", "Apply migrations") + "\n")
	sb.WriteString(row(5, "goforge app serve", "Start the server") + "\n")

	sb.WriteString("\n" + sectionStyle.Render("── Adding a feature") + "\n\n")
	sb.WriteString(row(6, "goforge gen:service <name>", "Scaffold service layer") + "\n")
	sb.WriteString(row(7, "goforge gen:event <name>", "Add event structs & listeners") + "\n")
	sb.WriteString(row(8, "goforge gen:migration <desc>", "New migration file") + "\n")
	sb.WriteString(row(9, "goforge migrate", "Apply it") + "\n")
	sb.WriteString(row(10, "goforge gen:proto", "Compile gRPC stubs") + "\n")

	sb.WriteString("\n" + sectionStyle.Render("── Optional tools") + "\n\n")
	sb.WriteString(row(11, "goforge gen:sqlc", "Type-safe SQL query layer") + "\n")
	sb.WriteString(row(12, "goforge gen:config <name>", "Add a config section") + "\n")
	sb.WriteString(row(13, "goforge gen:command <name>", "Custom console command") + "\n")

	sb.WriteString("\n" + sectionStyle.Render("── Cleanup") + "\n\n")
	sb.WriteString(row(14, "goforge rem:service <name>", "Remove a service") + "\n")
	sb.WriteString(row(15, "goforge rem:migration", "Remove latest migration") + "\n")
	sb.WriteString(row(16, "goforge rem:sqlc", "Remove SQLC integration") + "\n")
	sb.WriteString(row(17, "goforge rem:key", "Clear APP_KEY from .env") + "\n")

	return sb.String()
}
