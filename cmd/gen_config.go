package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmycin/GoForge/internal/env"
	"github.com/mmycin/GoForge/internal/scaffold"
	"github.com/mmycin/GoForge/internal/tui"
	"github.com/mmycin/GoForge/internal/tui/progress"
	"github.com/mmycin/GoForge/internal/tui/result"
	"github.com/mmycin/GoForge/internal/usecase"
	"github.com/spf13/cobra"
)

func newGenConfigCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "gen:config [name]",
		Short: "Add a new config section",
		Long: `Scaffold a new config file in core/config/ and wire it into AllConfig.

Example:
  goforge gen:config mailer
  → creates core/config/mailer.go with MailerConfig struct
  → wires it into AllConfig, the var block, and Load() in config.go`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			return runGenConfig(d, name)
		},
	}
}

func runGenConfig(d *Deps, name string) error {
	if name == "" {
		var err error
		name, err = promptInput(
			"Config Section Name",
			"snake_case. Creates core/config/<name>.go and wires it into AllConfig.",
			"mailer",
		)
		if err != nil {
			return err
		}
		if name == "" {
			tui.Info("Cancelled.")
			return nil
		}
	}
	name = strings.ToLower(strings.TrimSpace(name))

	cfg, _ := env.Load()
	modulePath := "github.com/mmycin/goforge"
	if cfg != nil && cfg.Module != "" {
		modulePath = cfg.Module
	}

	steps := []string{
		fmt.Sprintf("Writing core/config/%s.go", name),
		"Wiring into core/config/config.go",
	}
	pm := progress.New(fmt.Sprintf("Creating config '%s'", name), steps)
	progressCh := make(chan any, 16)
	var runErr error
	var finalFiles []usecase.GeneratedFile

	go func() {
		runErr = d.GenConfig.Run(name, modulePath, progressCh)
		close(progressCh)
	}()

	runner := newCollectingProgressRunner(pm, progressCh, &finalFiles)
	prog := tea.NewProgram(runner, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return err
	}

	if runErr != nil {
		tui.Error("gen:config failed: %v", runErr)
		return runErr
	}

	envPrefix := scaffold.EnvPrefix(name)
	nextStep := fmt.Sprintf("Add to .env:\n    %s_ENABLED=false\n    %s_HOST=localhost\n    %s_PORT=0",
		envPrefix, envPrefix, envPrefix)

	rm := result.New(fmt.Sprintf("Config section '%s' created", name), finalFiles, nextStep)
	rp := tea.NewProgram(rm, tea.WithAltScreen())
	_, err := rp.Run()
	return err
}
