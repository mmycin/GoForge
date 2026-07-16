package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmycin/GoForge/internal/tui"
	"github.com/mmycin/GoForge/internal/tui/confirm"
	"github.com/mmycin/GoForge/internal/tui/progress"
	"github.com/mmycin/GoForge/internal/tui/result"
	"github.com/mmycin/GoForge/internal/usecase"
	"github.com/spf13/cobra"
)

func newGenCommandCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "gen:command [name]",
		Short: "Create a custom console command",
		Long:  `Generate a new CLI command file in internal/console/. Run it via: goforge app run <name>.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			return runGenCommand(d, name)
		},
	}
}

func runGenCommand(d *Deps, name string) error {
	if name == "" {
		var err error
		name, err = promptInput(
			"Command Name",
			"Use colons for namespacing, e.g. emails:send-digest\nRun it via: goforge app run <name>",
			"emails:send-digest",
		)
		if err != nil {
			return err
		}
		if name == "" {
			tui.Info("Cancelled.")
			return nil
		}
	}

	steps := []string{"Writing command file"}
	pm := progress.New(fmt.Sprintf("Creating command '%s'", name), steps)
	progressCh := make(chan any, 16)
	var runErr error
	var finalFiles []usecase.GeneratedFile

	go func() {
		runErr = d.GenCommand.Run(name, progressCh)
		close(progressCh)
	}()

	runner := newCollectingProgressRunner(pm, progressCh, &finalFiles)
	prog := tea.NewProgram(runner, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return err
	}

	if runErr != nil {
		tui.Error("gen:command failed: %v", runErr)
		return runErr
	}

	rm := result.New(
		fmt.Sprintf("Command '%s' created", name),
		finalFiles,
		fmt.Sprintf("goforge app run %s", name),
	)
	rp := tea.NewProgram(rm, tea.WithAltScreen())
	_, err := rp.Run()
	return err
}

func newRemCommandCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rem:command [name]",
		Short: "Delete a custom console command",
		Long:  `Permanently delete the command file at internal/console/<name>_cmd.go.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemCommand(d, args[0])
		},
	}
}

func runRemCommand(d *Deps, name string) error {
	body := fmt.Sprintf(
		"This will permanently delete:\n\n  • internal/console/%s_cmd.go\n\nThis cannot be undone.",
		name,
	)

	cm := confirm.New("Remove Command  "+name, body, true)
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

	steps := []string{"Deleting command file"}
	pm := progress.New(fmt.Sprintf("Removing command '%s'", name), steps)
	progressCh := make(chan any, 16)
	var runErr error

	go func() {
		runErr = d.RemCommand.Run(name, progressCh)
		close(progressCh)
	}()

	prog := tea.NewProgram(newProgressRunner(pm, progressCh), tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return err
	}
	if runErr != nil {
		tui.Error("rem:command failed: %v", runErr)
	}
	return runErr
}
