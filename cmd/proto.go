package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmycin/GoForge/internal/env"
	"github.com/mmycin/GoForge/internal/tui"
	"github.com/mmycin/GoForge/internal/tui/confirm"
	"github.com/mmycin/GoForge/internal/tui/picker"
	"github.com/mmycin/GoForge/internal/tui/progress"
	"github.com/spf13/cobra"
)

func newGenProtoCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "gen:proto [service]",
		Short: "Compile proto files & generate gRPC stubs",
		Long:  `Compile .proto files with protoc and scaffold Go gRPC server stubs.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := ""
			if len(args) == 1 {
				svc = args[0]
			}
			return runGenProto(d, svc)
		},
	}
}

func runGenProto(d *Deps, serviceName string) error {
	// ── Service picker ────────────────────────────────────────────────────────
	if serviceName == "" {
		services, err := d.Generator.ListServices()
		if err != nil {
			return fmt.Errorf("could not list services: %w", err)
		}
		if len(services) == 0 {
			tui.Error("No services found — run goforge gen:service <name> first.")
			return nil
		}

		const allLabel = "All services"
		values := append([]string{allLabel}, services...)
		descs := make([]string, len(values))
		descs[0] = "Compile every proto file"
		for i, s := range services {
			descs[i+1] = "internal/proto/" + s
		}

		pm := picker.New("Compile Proto — select a service", values, descs)
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
		if fm.Selected() != allLabel {
			serviceName = fm.Selected()
		}
	}

	cfg, _ := env.Load()
	modulePath := "github.com/mmycin/goforge"
	if cfg != nil && cfg.Module != "" {
		modulePath = cfg.Module
	}

	title := "Compiling all proto files"
	if serviceName != "" {
		title = fmt.Sprintf("Compiling %s.proto", serviceName)
	}

	// gen:proto runs protoc (external subprocess) — use viewport for live output.
	return runWithViewport(d, title, func(ch chan<- any) error {
		return d.GenProto.Run(serviceName, modulePath, ch)
	})
}

// ── rem:proto ─────────────────────────────────────────────────────────────────

func newRemProtoCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rem:proto",
		Short: "Remove generated gRPC code",
		Long:  `Permanently remove generated .pb.go files and gen/ directories.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemProto(d)
		},
	}
}

func runRemProto(d *Deps) error {
	body := "This will remove all *.pb.go files and internal/proto/*/gen/ directories.\n\nThis cannot be undone."
	cm := confirm.New("Remove Generated Proto Code", body, true)
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

	// Use a progress screen so there's visible feedback.
	steps := []string{"Removing generated proto files"}
	pm := progress.New("Removing Generated Proto Code", steps)
	progressCh := make(chan any, 32)
	var runErr error

	go func() {
		runErr = d.RemProto.Run(progressCh)
		close(progressCh)
	}()

	prog := tea.NewProgram(newProgressRunner(pm, progressCh), tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return err
	}
	if runErr != nil {
		tui.Error("rem:proto failed: %v", runErr)
	}
	return runErr
}
