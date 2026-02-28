package cmd

import (
	"os"
	"path/filepath"

	"github.com/mmycin/GoForge/internal/env"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(removeServiceCmd)
}

var removeServiceCmd = &cobra.Command{
	Use:   "rem:service [name]",
	Short: "Remove an existing service",
	Long:  `Permanently remove a service, including its directory, proto files, and kernel registration.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		Info("Removing service: %s", name)
		removeService(name)
	},
}

func removeService(name string) {
	servicesDir := filepath.Join("internal/services", name)
	protoDir := filepath.Join("proto", name)

	// Check if service exists
	if _, err := os.Stat(servicesDir); os.IsNotExist(err) {
		ErrorLog("Service '%s' does not exist", name)
		os.Exit(1)
	}

	// Remove internal/services/<name>
	if err := os.RemoveAll(servicesDir); err != nil {
		ErrorLog("Failed to remove service directory: %v", err)
		os.Exit(1)
	}

	// Remove proto/<name>
	if err := os.RemoveAll(protoDir); err != nil {
		ErrorLog("Failed to remove proto directory: %v", err)
	}

	// Remove from kernel
	removeFromKernel(name)

	Success("Service '%s' removed successfully", name)
}

func removeFromKernel(name string) {
	cfg, err := env.Load()
	moduleName := "github.com/mmycin/goforge"
	if err == nil && cfg.Module != "" {
		moduleName = cfg.Module
	}

	if err := registerModels(moduleName); err != nil {
		Warning("Could not automatically update kernel.go: %v", err)
	}
}
