package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmycin/GoForge/internal/env"
	"github.com/mmycin/GoForge/internal/tui"
	"github.com/mmycin/GoForge/internal/tui/progress"
	"github.com/mmycin/GoForge/internal/tui/result"
	"github.com/mmycin/GoForge/internal/usecase"
	"github.com/spf13/cobra"
)

func newGenServiceCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "gen:service [name]",
		Short: "Scaffold a new service layer",
		Long:  `Generate a new service with handler, model, routes, docs, proto, and gRPC stub files.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			return runGenService(d, name)
		},
	}
}

func runGenService(d *Deps, name string) error {
	// Collect name via TUI if not supplied.
	if name == "" {
		var input string
		form := huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("Service Name").
				Description("Use snake_case. Files go to internal/services/<name>/").
				Placeholder("product_catalog").
				Value(&input).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("service name is required")
					}
					return nil
				}),
		)).WithTheme(huh.ThemeCatppuccin())

		p := tea.NewProgram(newWizardRunner(form, "New Service", 1), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return err
		}
		if form.State == huh.StateAborted {
			tui.Info("Cancelled.")
			return nil
		}
		name = strings.TrimSpace(input)
	}

	if name == "" {
		tui.Info("Cancelled.")
		return nil
	}

	cfg, _ := env.Load()
	modulePath := "github.com/mmycin/goforge"
	if cfg != nil && cfg.Module != "" {
		modulePath = cfg.Module
	}

	steps := []string{
		"Writing service.go", "Writing handler.go", "Writing model.go",
		"Writing routes.go", "Writing docs.go", "Writing grpc.go",
		"Writing proto file", "Updating kernel.go",
	}
	pm := progress.New(fmt.Sprintf("Scaffolding service '%s'", name), steps)
	progressCh := make(chan any, 64)

	// Collect generated files while draining the channel in the runner.
	// Use a pointer so the progress runner closure can write to it safely
	// (the runner runs on the tea goroutine, the use-case on a separate goroutine,
	// but the files pointer is only written from UseCaseDone which happens before
	// channel close — so reading after p.Run() is safe).
	var finalFiles []usecase.GeneratedFile
	var runErr error

	go func() {
		runErr = d.GenService.Run(name, modulePath, progressCh)
		close(progressCh)
	}()

	runner := newCollectingProgressRunner(pm, progressCh, &finalFiles)
	prog := tea.NewProgram(runner, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return err
	}

	if runErr != nil {
		tui.Error("gen:service failed: %v", runErr)
		return runErr
	}

	nextStep := fmt.Sprintf("goforge gen:migration add_%s_table", name)
	rm := result.New(fmt.Sprintf("Service '%s' scaffolded", name), finalFiles, nextStep)
	rp := tea.NewProgram(rm, tea.WithAltScreen())
	_, err := rp.Run()
	return err
}

// collectingProgressRunner is like progressRunner but captures GeneratedFiles
// from UseCaseDone into the provided pointer.
type collectingProgressRunner struct {
	progressRunner
	files *[]usecase.GeneratedFile
}

func newCollectingProgressRunner(pm progress.Model, ch <-chan any, files *[]usecase.GeneratedFile) collectingProgressRunner {
	return collectingProgressRunner{
		progressRunner: newProgressRunner(pm, ch),
		files:          files,
	}
}

func (r collectingProgressRunner) Init() tea.Cmd {
	return r.progressRunner.Init()
}

func (r collectingProgressRunner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Intercept UseCaseDone to capture files before forwarding.
	if tick, ok := msg.(progressTickMsg); ok {
		if done, ok := tick.event.(usecase.UseCaseDone); ok {
			*r.files = done.Files
		}
	}
	updated, cmd := r.progressRunner.Update(msg)
	// Re-wrap so our type is preserved in the final model.
	if pr, ok := updated.(progressRunner); ok {
		r.progressRunner = pr
	}
	return r, cmd
}

func (r collectingProgressRunner) View() string {
	return r.progressRunner.View()
}
