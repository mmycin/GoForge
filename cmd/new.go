package cmd

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new GoForge project",
	Long:  `Create a new GoForge project from the official template, with custom project and module names.`,
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Name of the Project: ")
		projectName, _ := reader.ReadString('\n')
		projectName = strings.TrimSpace(projectName)

		if projectName == "" {
			ErrorLog("Project name cannot be empty")
			return
		}

		fmt.Print("Name of the Module: ")
		moduleName, _ := reader.ReadString('\n')
		moduleName = strings.TrimSpace(moduleName)

		if moduleName == "" {
			ErrorLog("Module name cannot be empty")
			return
		}

		createNewProject(projectName, moduleName)
	},
}

func createNewProject(projectName, moduleName string) {
	Info("Creating new project: %s", projectName)

	// 1. Clone the template
	// Info("Cloning template from github.com/mmycin/goforge-template...")
	cloneCmd := exec.Command("git", "clone", "--branch", "main", "https://github.com/mmycin/goforge-template", projectName)
	if err := cloneCmd.Run(); err != nil {
		ErrorLog("Failed to clone template: %v", err)
		return
	}

	// 2. Remove .git directory to start fresh
	gitDir := filepath.Join(projectName, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		Warning("Failed to remove .git directory: %v", err)
	}

	// 3. Replace module name in all files
	// Info("Replacing module name 'github.com/mmycin/goforge' with '%s'...", moduleName)
	oldModule := "github.com/mmycin/goforge"

	err := filepath.WalkDir(projectName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Check if it contains the old module path
		if strings.Contains(string(content), oldModule) {
			newContent := strings.ReplaceAll(string(content), oldModule, moduleName)
			err = os.WriteFile(path, []byte(newContent), 0644)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		ErrorLog("Failed to replace module name: %v", err)
		return
	}

	// 4. Run go mod tidy
	Info("Running go mod tidy in %s...", projectName)
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = projectName
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		Warning("Failed to run go mod tidy: %v", err)
	}

	Success("Successfully created project %s!", projectName)
	Info("To get started:")
	Info("  cd %s", projectName)
	Info("  GoForge app serve")
}
