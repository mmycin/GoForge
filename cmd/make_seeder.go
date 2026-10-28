package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmycin/GoForge/internal/tui"
	"github.com/mmycin/GoForge/internal/tui/confirm"
	"github.com/mmycin/GoForge/internal/tui/progress"
	"github.com/mmycin/GoForge/internal/usecase"
	"github.com/spf13/cobra"
)

// ── gen:seeder ────────────────────────────────────────────────────────────────

func newGenSeederCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "gen:seeder [name]",
		Short: "Generate a new database seeder",
		Long:  `Scaffold a new seeder file under internal/database/seeders/. Register it in DatabaseSeeder.Run() to include it in the default seed run.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			return runGenSeeder(d, name)
		},
	}
}

func runGenSeeder(d *Deps, name string) error {
	if name == "" {
		var err error
		name, err = promptInput(
			"Seeder Name",
			"Use snake_case or PascalCase, e.g. user or UserSeeder",
			"user",
		)
		if err != nil {
			return err
		}
		if name == "" {
			tui.Info("Cancelled.")
			return nil
		}
	}

	steps := []string{"Writing seeder file"}
	pm := progress.New(fmt.Sprintf("Creating seeder '%s'", name), steps)
	progressCh := make(chan any, 16)
	var runErr error
	var finalFiles []usecase.GeneratedFile

	go func() {
		runErr = d.GenSeeder.Run(name, progressCh)
		close(progressCh)
	}()

	runner := newCollectingProgressRunner(pm, progressCh, &finalFiles)
	prog := tea.NewProgram(runner, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return err
	}

	if runErr != nil {
		tui.Error("gen:seeder failed: %v", runErr)
		return runErr
	}

	// Show a result card with next steps.
	tui.Success("Seeder created → %s", finalFiles[0].Path)
	tui.Info("Next steps:")
	tui.Info("  1. Open %s and add your seeding logic.", finalFiles[0].Path)
	tui.Info("  2. Register &%sSeeder{} in internal/database/seeders/database_seeder.go → DatabaseSeeder.Run().", name)
	tui.Info("  3. Run: goforge db:seed")
	return nil
}

// ── rem:seeder ────────────────────────────────────────────────────────────────

func newRemSeederCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rem:seeder [name]",
		Short: "Delete a database seeder file",
		Long:  `Permanently delete the seeder file at internal/database/seeders/<name>_seeder.go.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemSeeder(d, args[0])
		},
	}
}

func runRemSeeder(d *Deps, name string) error {
	body := fmt.Sprintf(
		"This will permanently delete:\n\n  • internal/database/seeders/%s_seeder.go\n\nRemember to also remove it from DatabaseSeeder.Run().\nThis cannot be undone.",
		name,
	)

	cm := confirm.New("Remove Seeder  "+name, body, true)
	cp := tea.NewProgram(cm, tea.WithAltScreen())
	finalModel, err := cp.Run()
	if err != nil {
		return err
	}
	fm, ok := finalModel.(confirm.Model)
	if !ok || !fm.Confirmed() {
		tui.Info("Cancelled.")
		return nil
	}

	steps := []string{"Deleting seeder file"}
	pm := progress.New(fmt.Sprintf("Removing seeder '%s'", name), steps)
	progressCh := make(chan any, 16)
	var runErr error

	go func() {
		runErr = d.RemSeeder.Run(name, progressCh)
		close(progressCh)
	}()

	prog := tea.NewProgram(newProgressRunner(pm, progressCh), tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return err
	}
	if runErr != nil {
		tui.Error("rem:seeder failed: %v", runErr)
	}
	return runErr
}
