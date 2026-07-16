package cmd

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

		selectedDBs := promptDatabases(reader)

		createNewProject(projectName, moduleName, selectedDBs)
	},
}

// allDrivers is the ordered list of supported databases shown in the prompt.
var allDrivers = []struct {
	label string // display name
	key   string // matches the driver file name (without .go)
}{
	{"SQLite", "sqlite"},
	{"MySQL", "mysql"},
	{"PostgreSQL", "postgresql"},
	{"SQL Server", "sqlserver"},
}

// driverFiles maps a driver key to its file path inside the cloned project.
var driverFiles = map[string]string{
	"sqlite":    "core/database/drivers/sqlite.go",
	"mysql":     "core/database/drivers/mysql.go",
	"postgresql": "core/database/drivers/postgresql.go",
	"sqlserver": "core/database/drivers/sqlserver.go",
}

// promptDatabases shows a numbered multi-select prompt and returns the keys
// of the databases the user chose. Defaults to SQLite (index 1) if the user
// presses Enter without typing anything.
func promptDatabases(reader *bufio.Reader) []string {
	fmt.Println("")
	fmt.Println("Select databases (comma-separated numbers, default: 1):")
	for i, d := range allDrivers {
		fmt.Printf("  %d) %s\n", i+1, d.label)
	}
	fmt.Print("> ")

	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)

	// Default: SQLite only
	if line == "" {
		return []string{"sqlite"}
	}

	seen := map[string]bool{}
	var selected []string

	for _, part := range strings.Split(line, ",") {
		part = strings.TrimSpace(part)
		n, err := strconv.Atoi(part)
		if err != nil || n < 1 || n > len(allDrivers) {
			Warning("Ignoring invalid selection %q", part)
			continue
		}
		key := allDrivers[n-1].key
		if !seen[key] {
			seen[key] = true
			selected = append(selected, key)
		}
	}

	if len(selected) == 0 {
		Warning("No valid databases selected, defaulting to SQLite")
		return []string{"sqlite"}
	}

	return selected
}

func createNewProject(projectName, moduleName string, selectedDBs []string) {
	Info("Creating new project: %s", projectName)

	// 1. Clone the template
	cloneCmd := exec.Command("git", "clone", "--branch", "main", "--quiet", "https://github.com/mmycin/goforge-template", projectName)
	cloneCmd.Stdout = nil
	cloneCmd.Stderr = nil
	if err := cloneCmd.Run(); err != nil {
		ErrorLog("Failed to clone template: %v", err)
		return
	}

	// 2. Remove .git directory to start fresh
	if err := os.RemoveAll(filepath.Join(projectName, ".git")); err != nil {
		Warning("Failed to remove .git directory: %v", err)
	}

	// 3. Replace module name in all Go and config files
	oldModule := "github.com/mmycin/goforge"
	err := filepath.WalkDir(projectName, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(content), oldModule) {
			return os.WriteFile(path, []byte(strings.ReplaceAll(string(content), oldModule, moduleName)), 0644)
		}
		return nil
	})
	if err != nil {
		ErrorLog("Failed to replace module name: %v", err)
		return
	}

	// 4. Delete driver files for unselected databases
	selectedSet := map[string]bool{}
	for _, k := range selectedDBs {
		selectedSet[k] = true
	}

	var removed []string
	for key, relPath := range driverFiles {
		if selectedSet[key] {
			continue
		}
		fullPath := filepath.Join(projectName, relPath)
		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			Warning("Could not delete driver file %s: %v", relPath, err)
			continue
		}
		removed = append(removed, key)
	}

	if len(removed) > 0 {
		Info("Removed unused database drivers: %s", strings.Join(removed, ", "))
		for _, key := range removed {
			Info("  ✓ Deleted %s", driverFiles[key])
		}
	}

	// 5. go mod tidy — drops packages for deleted drivers automatically
	Info("Running go mod tidy in %s...", projectName)
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = projectName
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		Warning("Failed to run go mod tidy: %v", err)
	}

	fmt.Println("")
	Success("Project %s created successfully!", projectName)
	Info("To get started:")
	Info("  cd %s", projectName)
	Info("  GoForge app serve")
}
