package usecase

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mmycin/GoForge/internal/env"
	"github.com/mmycin/GoForge/internal/infra"
)

const atlasConfigURI = "file://core/database/atlas.hcl"

// MigrateUseCase applies pending database migrations.
type MigrateUseCase struct {
	fs   infra.FileSystem
	exec infra.Executor
}

// NewMigrateUseCase constructs a MigrateUseCase.
func NewMigrateUseCase(fs infra.FileSystem, exec infra.Executor) *MigrateUseCase {
	return &MigrateUseCase{fs: fs, exec: exec}
}

// MigrateInfo contains the detected configuration shown in the pre-flight card.
type MigrateInfo struct {
	Migrator   string
	Connection string
	DBName     string
	DBHost     string
	DBPort     string
}

// Detect reads the project's environment and returns migration metadata.
func (uc *MigrateUseCase) Detect() MigrateInfo {
	cfg, _ := env.Load()
	info := MigrateInfo{Migrator: "atlas", Connection: "sqlite"}
	if cfg != nil {
		if cfg.DBMigrator != "" {
			info.Migrator = cfg.DBMigrator
		}
		if cfg.DBConnection != "" {
			info.Connection = cfg.DBConnection
		}
		info.DBName = cfg.DBName
		info.DBHost = cfg.DBHost
		info.DBPort = cfg.DBPort
	}
	return info
}

// Run applies migrations, streaming output to progress.
func (uc *MigrateUseCase) Run(progress Progress) error {
	cfg, _ := env.Load()

	migrator := "atlas"
	if cfg != nil && cfg.DBMigrator != "" {
		migrator = cfg.DBMigrator
	}
	send(progress, LogLine{Level: "info", Text: fmt.Sprintf("Migrator: %s", migrator)})

	if migrator == "gorm" {
		return uc.runGorm(cfg, progress)
	}
	return uc.runAtlas(cfg, "apply", "", progress)
}

func (uc *MigrateUseCase) runAtlas(cfg *env.Config, subCmd, migrationName string, progress Progress) error {
	// Pre-check: is atlas available?
	if _, err := exec.LookPath("atlas"); err != nil {
		send(progress, LogLine{Level: "error", Text: "atlas is not installed or not in PATH."})
		send(progress, LogLine{Level: "info", Text: "Install: https://atlasgo.io/docs#installation"})
		send(progress, UseCaseDone{})
		return fmt.Errorf("atlas not found")
	}

	atlasEnv := buildAtlasEnv(cfg)

	args := []string{"migrate", subCmd, "--env", "gorm", "--config", atlasConfigURI}
	if migrationName != "" {
		args = append(args, migrationName)
	}

	writer := &progressWriter{progress: progress}
	send(progress, StepStarted{Label: fmt.Sprintf("atlas migrate %s", subCmd)})
	if err := uc.exec.RunStreamedWithEnv(context.Background(), atlasEnv, writer, "atlas", args...); err != nil {
		writer.Flush()
		send(progress, StepFailed{Label: fmt.Sprintf("atlas migrate %s", subCmd), Err: err})
		return err
	}
	writer.Flush()
	send(progress, StepDone{Label: fmt.Sprintf("atlas migrate %s completed", subCmd)})
	send(progress, UseCaseDone{})
	return nil
}

func (uc *MigrateUseCase) runGorm(cfg *env.Config, progress Progress) error {
	modulePath := "github.com/mmycin/goforge"
	if cfg != nil && cfg.Module != "" {
		modulePath = cfg.Module
	}

	src := fmt.Sprintf(`package main
import (
	"fmt"
	"os"
	"%s/core/database"
	"%s/internal/services"
)
func main() {
	if err := database.Connect(); err != nil {
		fmt.Fprintln(os.Stderr, "connection failed:", err)
		os.Exit(1)
	}
	if err := database.DB.Gorm.AutoMigrate(services.Model()...); err != nil {
		fmt.Fprintln(os.Stderr, "migrate failed:", err)
		os.Exit(1)
	}
	fmt.Println("Migration completed")
}`, modulePath, modulePath)

	tmp := "goforge_tmp_migrate.go"
	if err := uc.fs.WriteFile(tmp, []byte(src), 0644); err != nil {
		return err
	}
	defer uc.fs.Remove(tmp)

	writer := &progressWriter{progress: progress}
	send(progress, StepStarted{Label: "Running GORM AutoMigrate"})
	if err := uc.exec.RunStreamed(context.Background(), writer, "go", "run", tmp); err != nil {
		send(progress, StepFailed{Label: "GORM AutoMigrate", Err: err})
		return err
	}
	send(progress, StepDone{Label: "GORM AutoMigrate completed"})
	send(progress, UseCaseDone{})
	return nil
}

func buildAtlasEnv(cfg *env.Config) []string {
	base := os.Environ()
	if cfg == nil {
		return base
	}
	add := func(key, val string) { base = append(base, key+"="+val) }
	add("DB_CONNECTION", cfg.DBConnection)
	add("DB_NAME", cfg.DBName)
	add("DB_USERNAME", cfg.DBUsername)
	add("DB_PASSWORD", cfg.DBPassword)
	add("DB_HOST", cfg.DBHost)
	add("DB_PORT", cfg.DBPort)
	add("DB_DEV_NAME", cfg.DBDevName)
	return base
}

// progressWriter bridges io.Writer to the progress channel line by line.
type progressWriter struct {
	progress Progress
	buf      strings.Builder
}

func (w *progressWriter) Write(p []byte) (int, error) {
	w.buf.Write(p)
	for {
		s := w.buf.String()
		idx := strings.IndexByte(s, '\n')
		if idx == -1 {
			break
		}
		line := s[:idx]
		w.buf.Reset()
		w.buf.WriteString(s[idx+1:])
		if strings.TrimSpace(line) != "" {
			send(w.progress, LogLine{Level: "info", Text: line})
		}
	}
	return len(p), nil
}

// Flush emits any remaining buffered content that didn't end with a newline.
func (w *progressWriter) Flush() {
	if s := strings.TrimSpace(w.buf.String()); s != "" {
		send(w.progress, LogLine{Level: "info", Text: s})
		w.buf.Reset()
	}
}
