package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmycin/GoForge/internal/env"
	"github.com/mmycin/GoForge/internal/tui"
	"github.com/mmycin/GoForge/internal/tui/picker"
	"github.com/mmycin/GoForge/internal/tui/progress"
	"github.com/mmycin/GoForge/internal/tui/result"
	"github.com/mmycin/GoForge/internal/usecase"
	"github.com/spf13/cobra"
)

func newGenEventCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "gen:event [service]",
		Short: "Generate event structs & listeners for a service",
		Long:  `Scaffold events.go and listeners.go for an existing service.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := ""
			if len(args) == 1 {
				svc = args[0]
			}
			return runGenEvent(d, svc)
		},
	}
}

func runGenEvent(d *Deps, serviceName string) error {
	// ── Service picker ────────────────────────────────────────────────────────
	if serviceName == "" {
		services, err := d.Generator.ListServices()
		if err != nil || len(services) == 0 {
			tui.Error("No services found — run goforge gen:service <name> first.")
			return nil
		}

		descs := make([]string, len(services))
		for i, s := range services {
			descs[i] = "internal/services/" + s
		}

		pm := picker.New("Generate Events — select a service", services, descs)
		pp := tea.NewProgram(pm, tea.WithAltScreen())
		finalModel, err := pp.Run()
		if err != nil {
			return err
		}
		fm, ok := finalModel.(picker.Model)
		if !ok || fm.Selected() == "" {
			tui.Info("Cancelled.")
			return nil
		}
		serviceName = fm.Selected()
	}

	cfg, _ := env.Load()
	modulePath := "github.com/mmycin/goforge"
	if cfg != nil && cfg.Module != "" {
		modulePath = cfg.Module
	}

	// ── Run with progress screen ──────────────────────────────────────────────
	// Even though gen:event is fast, always show a progress screen so the user
	// sees what happened, and avoid the "silent drain then immediate result"
	// pattern that causes terminal state issues.
	steps := []string{
		"Writing events.go",
		"Writing listeners.go",
	}
	pm2 := progress.New(fmt.Sprintf("Generating events for '%s'", serviceName), steps)
	progressCh := make(chan any, 32)
	var runErr error
	var finalFiles []usecase.GeneratedFile

	go func() {
		runErr = d.GenEvent.Run(serviceName, modulePath, progressCh)
		close(progressCh)
	}()

	// Use collectingProgressRunner to capture files safely on the tea goroutine.
	runner := newCollectingProgressRunner(pm2, progressCh, &finalFiles)
	prog := tea.NewProgram(runner, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return err
	}

	if runErr != nil {
		tui.Error("gen:event failed: %v", runErr)
		return runErr
	}

	rm := result.New(
		fmt.Sprintf("Event files generated for '%s'", serviceName),
		finalFiles,
		"Register your listeners in the service boot file.",
	)
	rp := tea.NewProgram(rm, tea.WithAltScreen())
	_, err := rp.Run()
	return err
}

// ── Shared input helpers ──────────────────────────────────────────────────────

func promptInput(title, description, placeholder string) (string, error) {
	var value string
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title(title).
			Description(description).
			Placeholder(placeholder).
			Value(&value).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("%s is required", title)
				}
				return nil
			}),
	)).WithTheme(huh.ThemeCatppuccin())

	p := tea.NewProgram(newWizardRunner(form, title, 1), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return "", err
	}
	if form.State == huh.StateAborted {
		return "", nil
	}
	return strings.TrimSpace(value), nil
}
