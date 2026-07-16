package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mmycin/GoForge/internal/env"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(genEventCmd)
}

var genEventCmd = &cobra.Command{
	Use:   "gen:event [service]",
	Short: "Generate events.go and listeners.go for a service",
	Long:  `Scaffold event structs and listener stubs for an existing service.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		service := args[0]
		Info("Generating event files for service: %s", service)
		genEventFiles(service)
	},
}

func genEventFiles(service string) {
	serviceDir := filepath.Join("internal", "services", service)

	// Guard: service directory must already exist
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		ErrorLog("Service '%s' does not exist at %s", service, serviceDir)
		ErrorLog("Run 'gen:service %s' first", service)
		os.Exit(1)
	}

	// Resolve module name from go.mod (falls back to default)
	moduleName := "github.com/mmycin/goforge"
	if cfg, err := env.Load(); err == nil && cfg.Module != "" {
		moduleName = cfg.Module
	} else {
		Warning("Could not read module from go.mod, using default: %s", moduleName)
	}

	camel := toCamelCase(service)

	writeEventFile(service, camel, serviceDir)
	writeListenerFile(service, camel, serviceDir, moduleName)
}

// writeEventFile generates events.go with topic constants and event structs.
func writeEventFile(service, camel, dir string) {
	target := filepath.Join(dir, "events.go")
	if _, err := os.Stat(target); err == nil {
		ErrorLog("events.go already exists in %s", dir)
		os.Exit(1)
	}

	content := fmt.Sprintf(`package %s

// Topic constants for %s events.
// Use these when registering listeners and dispatching events.
const (
	Event%sCreated = "%s.created"
	Event%sUpdated = "%s.updated"
	Event%sDeleted = "%s.deleted"
)

// %sCreated is fired after a %s is successfully inserted.
type %sCreated struct {
	ID uint
}

// %sUpdated is fired after a %s is updated.
type %sUpdated struct {
	ID uint
}

// %sDeleted is fired after a %s is removed.
type %sDeleted struct {
	ID uint
}
`,
		service,
		service,
		camel, service,
		camel, service,
		camel, service,
		camel, service,
		camel,
		camel, service,
		camel,
		camel, service,
		camel,
	)

	if err := os.WriteFile(target, []byte(content), 0644); err != nil {
		ErrorLog("Failed to write events.go: %v", err)
		os.Exit(1)
	}
	Success("Generated %s", target)
}

// writeListenerFile generates listeners.go with stub listener structs and compile-time assertions.
func writeListenerFile(service, camel, dir, moduleName string) {
	target := filepath.Join(dir, "listeners.go")
	if _, err := os.Stat(target); err == nil {
		ErrorLog("listeners.go already exists in %s", dir)
		os.Exit(1)
	}

	content := fmt.Sprintf(`package %s

import (
	"context"
	"fmt"

	"%s/core/database"
	coreevents "%s/core/events"
)

// AuditOn%sCreated handles the %sCreated event.
type AuditOn%sCreated struct {
	DB *database.Database
}

func (l *AuditOn%sCreated) Handle(ctx context.Context, e %sCreated) error {
	fmt.Printf("%%s created: %%d\n", "%s", e.ID)
	// TODO: implement handler logic
	return nil
}

// AuditOn%sUpdated handles the %sUpdated event.
type AuditOn%sUpdated struct {
	DB *database.Database
}

func (l *AuditOn%sUpdated) Handle(ctx context.Context, e %sUpdated) error {
	fmt.Printf("%%s updated: %%d\n", "%s", e.ID)
	// TODO: implement handler logic
	return nil
}

// AuditOn%sDeleted handles the %sDeleted event.
type AuditOn%sDeleted struct {
	DB *database.Database
}

func (l *AuditOn%sDeleted) Handle(ctx context.Context, e %sDeleted) error {
	fmt.Printf("%%s deleted: %%d\n", "%s", e.ID)
	// TODO: implement handler logic
	return nil
}

// Compile-time assertions: ensure all listeners satisfy the Listener[T] interface.
var (
	_ coreevents.Listener[%sCreated] = (*AuditOn%sCreated)(nil)
	_ coreevents.Listener[%sUpdated] = (*AuditOn%sUpdated)(nil)
	_ coreevents.Listener[%sDeleted] = (*AuditOn%sDeleted)(nil)
)
`,
		// package
		service,
		// imports
		moduleName,
		moduleName,
		// AuditOnCreated
		camel, camel,
		camel,
		camel, camel,
		service,
		// AuditOnUpdated
		camel, camel,
		camel,
		camel, camel,
		service,
		// AuditOnDeleted
		camel, camel,
		camel,
		camel, camel,
		service,
		// compile-time assertions
		camel, camel,
		camel, camel,
		camel, camel,
	)

	if err := os.WriteFile(target, []byte(content), 0644); err != nil {
		ErrorLog("Failed to write listeners.go: %v", err)
		os.Exit(1)
	}
	Success("Generated %s", target)
}
