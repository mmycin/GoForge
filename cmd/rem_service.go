package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmycin/GoForge/internal/env"
	"github.com/mmycin/GoForge/internal/tui"
	"github.com/mmycin/GoForge/internal/tui/confirm"
	"github.com/mmycin/GoForge/internal/tui/progress"
	"github.com/spf13/cobra"
)

func newRemServiceCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rem:service [name]",
		Short: "Permanently remove a service",
		Long:  `Remove a service directory, its proto files, and its kernel.go registration.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemService(d, args[0])
		},
	}
}

func runRemService(d *Deps, name string) error {
	body := fmt.Sprintf(
		"You are about to permanently remove:\n\n"+
			"  • internal/services/%s/  (all files)\n"+
			"  • internal/proto/%s/     (all files)\n"+
			"  • Registration in internal/services/kernel.go\n\n"+
			"This cannot be undone.",
		name, name,
	)

	cm := confirm.New("Remove Service  "+name, body, true)
	cp := tea.NewProgram(cm, tea.WithAltScreen())
	finalModel, err := cp.Run()
	if err != nil {
		return err
	}
	fm, ok := finalModel.(confirm.Model)
	if !ok || !fm.Confirmed() {
		tui.Info("Cancelled — service was not removed.")
		return nil
	}

	cfg, _ := env.Load()
	modulePath := "github.com/mmycin/goforge"
	if cfg != nil && cfg.Module != "" {
		modulePath = cfg.Module
	}

	steps := []string{
		fmt.Sprintf("Removing internal/services/%s", name),
		fmt.Sprintf("Removing internal/proto/%s", name),
		"Updating kernel.go",
	}
	pm := progress.New(fmt.Sprintf("Removing service '%s'", name), steps)
	progressCh := make(chan any, 32)

	go func() {
		_ = d.RemService.Run(name, modulePath, progressCh)
		close(progressCh)
	}()

	prog := tea.NewProgram(newProgressRunner(pm, progressCh), tea.WithAltScreen())
	_, err = prog.Run()
	return err
}
