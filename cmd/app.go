package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(appCmd)
	appCmd.AddCommand(appServeCmd)
}

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Manage and run application-specific commands",
	Long:  `Proxy to the local application. Commands inside app will run locally against the user's main.go.`,
}

var appServeCmd = &cobra.Command{
	Use:   "serve [command]",
	Short: "Run the application or a locally defined GoForge task",
	Long:  `Executes the specified command context inside the current GoForge project using go run cmd/main.go. Defaults to 'serve'.`,
	// We want to accept any number of arguments after `serve`
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		Info("Running application command...")

		var execArgs []string
		if len(args) == 0 {
			execArgs = []string{"run", "cmd/main.go", "serve"}
		} else {
			execArgs = append([]string{"run", "cmd/main.go"}, args...)
		}

		proxy := exec.Command("go", execArgs...)
		proxy.Stdout = os.Stdout
		proxy.Stderr = os.Stderr
		proxy.Stdin = os.Stdin

		if err := proxy.Run(); err != nil {
			ErrorLog("Command failed: %v", err)
			os.Exit(1)
		}
	},
}
