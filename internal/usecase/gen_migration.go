package usecase

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mmycin/GoForge/internal/env"
	"github.com/mmycin/GoForge/internal/infra"
)

// GenMigrationUseCase creates a new Atlas migration file.
type GenMigrationUseCase struct {
	fs   infra.FileSystem
	exec infra.Executor
}

// NewGenMigrationUseCase constructs a GenMigrationUseCase.
func NewGenMigrationUseCase(fs infra.FileSystem, exec infra.Executor) *GenMigrationUseCase {
	return &GenMigrationUseCase{fs: fs, exec: exec}
}

// Run generates a new migration with the given name.
func (uc *GenMigrationUseCase) Run(name string, progress Progress) error {
	// Pre-check: is atlas available?
	if _, err := exec.LookPath("atlas"); err != nil {
		send(progress, LogLine{Level: "error", Text: "atlas is not installed or not in PATH."})
		send(progress, LogLine{Level: "info", Text: "Install: https://atlasgo.io/docs#installation"})
		send(progress, UseCaseDone{})
		return fmt.Errorf("atlas not found: install from https://atlasgo.io/docs#installation")
	}

	cfg, _ := env.Load()
	atlasEnv := buildAtlasEnv(cfg)
	writer := &progressWriter{progress: progress}

	// ── atlas migrate diff ────────────────────────────────────────────────────
	send(progress, StepStarted{Label: "Running atlas migrate diff"})
	if err := uc.runAtlasStreamed(atlasEnv, writer, progress,
		"migrate", "diff", "--env", "gorm", "--config", atlasConfigURI, name,
	); err != nil {
		send(progress, StepFailed{Label: "atlas migrate diff", Err: err})
		uc.printAtlasTip(cfg, progress)
		send(progress, UseCaseDone{})
		return fmt.Errorf("atlas migrate diff: %w", err)
	}
	send(progress, StepDone{Label: "Migration file created"})

	// ── Clean SQL ─────────────────────────────────────────────────────────────
	send(progress, StepStarted{Label: "Cleaning up SQL files"})
	uc.cleanupSQL()
	send(progress, StepDone{Label: "SQL files cleaned"})

	// ── atlas migrate hash ────────────────────────────────────────────────────
	send(progress, StepStarted{Label: "Updating atlas hash"})
	if err := uc.runAtlasStreamed(atlasEnv, writer, progress,
		"migrate", "hash", "--env", "gorm", "--config", atlasConfigURI,
	); err != nil {
		send(progress, StepFailed{Label: "atlas migrate hash", Err: err})
		send(progress, UseCaseDone{})
		return fmt.Errorf("atlas migrate hash: %w", err)
	}
	send(progress, StepDone{Label: "Migration hash updated"})

	send(progress, UseCaseDone{})
	return nil
}

// runAtlasStreamed runs atlas, streaming output to the progress writer.
// On failure it also captures any buffered stderr and emits it as log lines.
func (uc *GenMigrationUseCase) runAtlasStreamed(
	atlasEnv []string, writer *progressWriter, progress Progress, args ...string,
) error {
	// Use a tee writer: stream to viewport AND capture for error reporting.
	var captured bytes.Buffer
	tee := &teeWriter{a: writer, b: &captured}

	err := uc.exec.RunStreamedWithEnv(context.Background(), atlasEnv, tee, "atlas", args...)
	if err != nil {
		// Flush any remaining buffered content.
		writer.Flush()
		// If the stream already wrote lines, they're visible. If not (e.g. atlas
		// printed to a real stderr instead), emit the captured bytes as log lines.
		if captured.Len() > 0 {
			for _, line := range strings.Split(captured.String(), "\n") {
				if strings.TrimSpace(line) != "" {
					send(progress, LogLine{Level: "error", Text: line})
				}
			}
		}
	} else {
		writer.Flush()
	}
	return err
}

// printAtlasTip sends a contextual hint when atlas migrate diff fails.
func (uc *GenMigrationUseCase) printAtlasTip(cfg *env.Config, progress Progress) {
	if cfg == nil {
		return
	}
	conn := cfg.DBConnection
	if (conn == "mysql" || conn == "postgres" || conn == "postgresql") && cfg.DBDevName == "" {
		send(progress, LogLine{Level: "warn", Text: ""})
		send(progress, LogLine{Level: "warn", Text: "Tip: Atlas requires a clean dev database for schema diffing."})
		send(progress, LogLine{Level: "info", Text: fmt.Sprintf(
			"1. Create an empty database: CREATE DATABASE %s_dev;", cfg.DBName,
		)})
		send(progress, LogLine{Level: "info", Text: fmt.Sprintf(
			"2. Add to .env:  DB_DEV_NAME=%s_dev", cfg.DBName,
		)})
		send(progress, LogLine{Level: "info", Text: "3. Run the command again."})
	}
}

func (uc *GenMigrationUseCase) cleanupSQL() {
	entries, err := uc.fs.ReadDir("internal/database/migrations")
	if err != nil {
		return
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		path := filepath.Join("internal/database/migrations", e.Name())
		content, err := uc.fs.ReadFile(path)
		if err != nil {
			continue
		}
		cleaned := strings.ReplaceAll(string(content), "`", "")
		lines := strings.Split(cleaned, "\n")
		for i, line := range lines {
			if idx := strings.Index(line, "COLLATE"); idx != -1 {
				trimmed := strings.TrimSpace(line[:idx])
				if !strings.HasSuffix(trimmed, ";") {
					trimmed += ";"
				}
				lines[i] = trimmed
			}
		}
		_ = uc.fs.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	}
}

// teeWriter writes to two io.Writers simultaneously.
type teeWriter struct {
	a, b interface{ Write([]byte) (int, error) }
}

func (t *teeWriter) Write(p []byte) (int, error) {
	n, err := t.a.Write(p)
	t.b.Write(p) //nolint:errcheck
	return n, err
}
