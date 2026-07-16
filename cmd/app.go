package cmd

import (
	"os"
	"os/exec"

	"github.com/mmycin/GoForge/internal/tui"
	"github.com/spf13/cobra"
)

func newAppCmd() *cobra.Command {
	appCmd := &cobra.Command{
		Use:   "app",
		Short: "Manage and run application-specific commands",
		Long:  `Proxy commands to the local GoForge application via go run app/main.go.`,
	}
	appCmd.AddCommand(newAppServeCmd())
	return appCmd
}

func newAppServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "serve [args...]",
		Short:              "Start the development server",
		Long:               `Run the application or a locally defined GoForge task using go run app/main.go.`,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			tui.Info("Starting application…")

			execArgs := []string{"run", "app/main.go"}
			if len(args) == 0 {
				execArgs = append(execArgs, "serve")
			} else {
				execArgs = append(execArgs, args...)
			}

			proxy := exec.Command("go", execArgs...)
			proxy.Stdout = os.Stdout
			proxy.Stderr = os.Stderr
			proxy.Stdin = os.Stdin

			if err := proxy.Run(); err != nil {
				tui.Error("Command failed: %v", err)
				return err
			}
			return nil
		},
	}
}
