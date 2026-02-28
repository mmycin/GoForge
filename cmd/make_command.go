package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(makeCommandCmd)
	rootCmd.AddCommand(removeCommandCmd)
}

var makeCommandCmd = &cobra.Command{
	Use:   "gen:command [name]",
	Short: "Create a new console command",
	Long:  `Generate a new CLI command to be executed via 'GoForge app run [name]'.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		Info("Creating custom command: %s", name)
		genCommandFile(name)
	},
}

func genCommandFile(name string) {
	// The directory in the target project where custom commands live
	cmdDir := filepath.Join("internal", "console")

	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		ErrorLog("Failed to assure internal/console exists: %v", err)
		os.Exit(1)
	}

	// name is e.g. "add:test"
	safeName := strings.ReplaceAll(name, ":", "_")
	safeName = strings.ReplaceAll(safeName, "-", "_")

	// camel struct
	camelName := toCamelCase(safeName) + "Cmd"

	fileName := safeName + "_cmd.go"
	targetPath := filepath.Join(cmdDir, fileName)

	if _, err := os.Stat(targetPath); err == nil {
		ErrorLog("Command file '%s' already exists at %s", name, targetPath)
		os.Exit(1)
	}

	content := fmt.Sprintf(`package console

import (
	"fmt"
	"github.com/spf13/cobra"
)

var %s = &cobra.Command{
	Use:   "%s",
	Short: "Description for %s",
	Long:  %sA comprehensive description for %s%s,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Executing custom command: %s")
		// TODO: Add your custom command logic here
	},
}

func init() {
	// Root command registration is handled automatically by init() in this package
	rootCmd.AddCommand(%s)
}
`, camelName, name, name, "`", name, "`", name, camelName)

	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		ErrorLog("Failed to write command file: %v", err)
		os.Exit(1)
	}

	Success("Generated new command at %s", targetPath)
	Info("You can execute it using: GoForge app run %s", name)
}

var removeCommandCmd = &cobra.Command{
	Use:   "rem:command [name]",
	Short: "Remove a generated console command",
	Long:  `Permanently delete the local file corresponding to a custom CLI command.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		Info("Removing custom command: %s", name)
		remCommandFile(name)
	},
}

func remCommandFile(name string) {
	cmdDir := filepath.Join("internal", "console")
	if _, err := os.Stat(cmdDir); os.IsNotExist(err) {
		ErrorLog("Directory internal/console does not exist")
		os.Exit(1)
	}

	safeName := strings.ReplaceAll(name, ":", "_")
	safeName = strings.ReplaceAll(safeName, "-", "_")

	fileName := safeName + "_cmd.go"
	targetPath := filepath.Join(cmdDir, fileName)

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		ErrorLog("Command file '%s' does not exist at %s", name, targetPath)
		os.Exit(1)
	}

	if err := os.Remove(targetPath); err != nil {
		ErrorLog("Failed to remove command file: %v", err)
		os.Exit(1)
	}

	Success("Removed command at %s", targetPath)
}
