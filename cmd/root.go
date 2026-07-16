package cmd

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmycin/GoForge/internal/tui"
	"github.com/mmycin/GoForge/internal/tui/dashboard"
	"github.com/spf13/cobra"
)

// deps is the single shared dependency graph for all commands.
var deps *Deps

// rootCmd is declared without RunE to break the initialization cycle.
// RunE is assigned in Execute() before cobra parses anything.
var rootCmd = &cobra.Command{
	Use:           "GoForge",
	Short:         "A production-ready Go application framework CLI",
	Long:          "GoForge is a comprehensive, production-ready Go framework CLI.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute is called from main.go.
func Execute() {
	deps = NewDeps()

	// Assign RunE here to avoid an init-time cycle.
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runDashboard()
	}

	registerCommands(deps)

	if err := rootCmd.Execute(); err != nil {
		tui.Error("%v", err)
		os.Exit(1)
	}
}

// registerCommands wires all cobra subcommands with the dependency graph.
func registerCommands(d *Deps) {
	rootCmd.AddCommand(
		newNewCmd(d),
		newAppCmd(),
		newVersionCmd(),
		newReadmeCmd(),
		newGenKeyCmd(d),
		newRemKeyCmd(d),
		newGenServiceCmd(d),
		newRemServiceCmd(d),
		newGenEventCmd(d),
		newGenCommandCmd(d),
		newRemCommandCmd(d),
		newGenConfigCmd(d),
		newGenProtoCmd(d),
		newRemProtoCmd(d),
		newMigrateCmd(d),
		newGenMigrationCmd(d),
		newRemMigrationCmd(d),
		newGenSqlcCmd(d),
		newRemSqlcCmd(d),
		newLoaderCmd(d),
	)
}

// runDashboard launches the interactive command browser. When the user selects
// a command, it re-invokes cobra with that command as the argument.
func runDashboard() error {
	m := dashboard.New()
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return err
	}

	fm, ok := final.(dashboard.Model)
	if !ok || fm.Selected().Name == "" {
		return nil // quit without selecting
	}

	// Re-run cobra with the selected command split into args.
	rootCmd.SetArgs(splitArgs(fm.Selected().Name))
	return rootCmd.Execute()
}

// splitArgs splits "app serve" → ["app", "serve"].
func splitArgs(s string) []string {
	var parts []string
	cur := ""
	for _, r := range s {
		if r == ' ' {
			if cur != "" {
				parts = append(parts, cur)
				cur = ""
			}
		} else {
			cur += string(r)
		}
	}
	if cur != "" {
		parts = append(parts, cur)
	}
	return parts
}
