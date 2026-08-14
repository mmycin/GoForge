package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmycin/GoForge/internal/tui"
	"github.com/mmycin/GoForge/internal/tui/confirm"
	"github.com/mmycin/GoForge/internal/tui/progress"
	"github.com/spf13/cobra"
)

func newRemGRPCCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "rem:grpc",
		Short: "Remove all gRPC infrastructure, leaving REST only",
		Long: `Permanently remove every gRPC artifact from this project:

  • boot/server/grpc.go
  • boot/client/grpc_client.go
  • boot/server/middleware/grpc.go
  • core/config/grpc.go
  • GRPC field from core/config/config.go
  • grpc.go stub from every service under internal/services/
  • internal/proto/ contents
  • internal/tests/grpc/ test files
  • gRPC wiring from internal/services/kernel.go
  • gRPC server construction from core/console/serve.go
  • Runs go mod tidy to drop unused gRPC dependencies

This cannot be undone.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemGRPC(d)
		},
	}
}

func runRemGRPC(d *Deps) error {
	body := "You are about to permanently remove ALL gRPC infrastructure:\n\n" +
		"  • boot/server/grpc.go\n" +
		"  • boot/client/grpc_client.go\n" +
		"  • boot/server/middleware/grpc.go\n" +
		"  • core/config/grpc.go  (+ GRPC field in config.go)\n" +
		"  • internal/services/*/grpc.go  (per-service stubs)\n" +
		"  • internal/proto/  (all proto files)\n" +
		"  • internal/tests/grpc/  (gRPC tests)\n" +
		"  • gRPC wiring in kernel.go and serve.go\n" +
		"  • go mod tidy  (drops google.golang.org/grpc + deps)\n\n" +
		"Only the REST server will remain.\n\n" +
		"This cannot be undone."

	cm := confirm.New("Remove gRPC", body, true)
	cp := tea.NewProgram(cm, tea.WithAltScreen())
	finalModel, err := cp.Run()
	if err != nil {
		return err
	}
	fm, ok := finalModel.(confirm.Model)
	if !ok || !fm.Confirmed() {
		tui.Info("Cancelled — gRPC was not removed.")
		return nil
	}

	steps := []string{
		"Removing boot/server/grpc.go",
		"Removing boot/client/grpc_client.go",
		"Removing boot/server/middleware/grpc.go",
		"Removing core/config/grpc.go",
		"Patching core/config/config.go",
		"Removing per-service grpc.go stubs",
		"Removing internal/proto contents",
		"Removing internal/tests/grpc",
		"Patching internal/services/kernel.go",
		"Patching core/console/serve.go",
		"Patching boot/kernel.go",
		"Running go mod tidy",
	}

	pm := progress.New("Removing gRPC", steps)
	progressCh := make(chan any, 64)

	go func() {
		_ = d.RemGRPC.Run(progressCh)
		close(progressCh)
	}()

	prog := tea.NewProgram(newProgressRunner(pm, progressCh), tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return err
	}

	tui.Success(fmt.Sprintf("gRPC removed. Your project is now REST-only."))
	return nil
}
