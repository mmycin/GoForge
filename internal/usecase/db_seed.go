package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/mmycin/GoForge/internal/env"
	"github.com/mmycin/GoForge/internal/infra"
)

// DbSeedUseCase runs the project's database seeders by generating a temporary
// Go program that imports the project's seeders package and executes them.
// This mirrors how the GORM AutoMigrate use-case works — no cobra command is
// needed inside the template project itself.
type DbSeedUseCase struct {
	fs   infra.FileSystem
	exec infra.Executor
}

// NewDbSeedUseCase constructs a DbSeedUseCase.
func NewDbSeedUseCase(fs infra.FileSystem, exec infra.Executor) *DbSeedUseCase {
	return &DbSeedUseCase{fs: fs, exec: exec}
}

// DbSeedInfo holds metadata shown in the pre-flight confirmation card.
type DbSeedInfo struct {
	Connection string
	DBName     string
	DBHost     string
	DBPort     string
}

// Detect reads project env and returns seed metadata for the confirmation UI.
func (uc *DbSeedUseCase) Detect() DbSeedInfo {
	cfg, _ := env.Load()
	info := DbSeedInfo{Connection: "sqlite"}
	if cfg != nil {
		if cfg.DBConnection != "" {
			info.Connection = cfg.DBConnection
		}
		info.DBName = cfg.DBName
		info.DBHost = cfg.DBHost
		info.DBPort = cfg.DBPort
	}
	return info
}

// Run executes the project seeders by generating and running a temporary Go program.
// className is the optional --class value; empty string means "DatabaseSeeder".
func (uc *DbSeedUseCase) Run(className string, progress Progress) error {
	cfg, _ := env.Load()

	modulePath := "github.com/mmycin/goforge"
	if cfg != nil && cfg.Module != "" {
		modulePath = cfg.Module
	}

	seederName := "DatabaseSeeder"
	if className != "" {
		seederName = className
	}

	label := fmt.Sprintf("db:seed (%s)", seederName)
	send(progress, StepStarted{Label: label})

	// Build a self-contained main program that:
	//   1. imports the project's seeder registry (triggering all init()s)
	//   2. connects to the DB using the project's own Connect()
	//   3. resolves the target seeder and calls Run()
	src := fmt.Sprintf(`package main

import (
	"fmt"
	"log"
	"os"

	"%s/core/config"
	"%s/core/database"
	_ "%s/internal/database/seeders"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}

	db, err := database.Connect(cfg.DB)
	if err != nil {
		fmt.Fprintln(os.Stderr, "db connection error:", err)
		os.Exit(1)
	}

	gormDB, err := db.GetGorm()
	if err != nil {
		fmt.Fprintln(os.Stderr, "gorm error:", err)
		os.Exit(1)
	}

	seederName := %q
	s, ok := database.ResolveSeeder(seederName)
	if !ok {
		log.Fatalf("seeder %%q not found — ensure it is registered in internal/database/seeders/registry.go", seederName)
	}

	fmt.Printf("→ Running %%s...\n", seederName)
	if err := s.Run(gormDB); err != nil {
		fmt.Fprintln(os.Stderr, "seed error:", err)
		os.Exit(1)
	}
	fmt.Printf("✓ %%s completed successfully\n", seederName)
}
`, modulePath, modulePath, modulePath, seederName)

	tmp := "goforge_tmp_seed.go"
	if err := uc.fs.WriteFile(tmp, []byte(src), 0644); err != nil {
		send(progress, StepFailed{Label: label, Err: err})
		return err
	}
	defer uc.fs.Remove(tmp)

	send(progress, LogLine{Level: "info", Text: fmt.Sprintf("Executing: %s", seederName)})

	writer := &progressWriter{progress: progress}
	if err := uc.exec.RunStreamed(context.Background(), writer, "go", "run", tmp); err != nil {
		writer.Flush()
		send(progress, StepFailed{Label: label, Err: err})
		return err
	}
	writer.Flush()
	send(progress, StepDone{Label: label + " completed"})
	send(progress, UseCaseDone{})
	return nil
}

// seederClassLabel returns a presentable label for the progress card.
func seederClassLabel(className string) string {
	if className == "" {
		return "DatabaseSeeder"
	}
	return strings.TrimSuffix(className, "Seeder") + "Seeder"
}
