package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(appCmd)
	appCmd.AddCommand(appRunCmd)
}

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Manage and run application-specific commands",
	Long:  `Proxy to the local application. Commands inside app will run locally against the user's main.go.`,
}

var appRunCmd = &cobra.Command{
	Use:   "run [command]",
	Short: "Run a locally defined GoForge task or custom command",
	Long:  `Executes the specified command context inside the current GoForge project using go run cmd/main.go.`,
	// We want to accept any number of arguments after `run`
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			ErrorLog("You must specify a command to run. E.g: GoForge app run my:cmd")
			os.Exit(1)
		}

		Info("Running application command: %v", args)

		execArgs := append([]string{"run", "cmd/main.go"}, args...)
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
