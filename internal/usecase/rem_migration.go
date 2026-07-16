package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mmycin/GoForge/internal/env"
	"github.com/mmycin/GoForge/internal/infra"
)

// RemMigrationUseCase removes the most recent migration file and updates the atlas hash.
type RemMigrationUseCase struct {
	fs   infra.FileSystem
	exec infra.Executor
}

// NewRemMigrationUseCase constructs a RemMigrationUseCase.
func NewRemMigrationUseCase(fs infra.FileSystem, exec infra.Executor) *RemMigrationUseCase {
	return &RemMigrationUseCase{fs: fs, exec: exec}
}

// Run removes the latest migration file and rehashes.
func (uc *RemMigrationUseCase) Run(progress Progress) error {
	migrationDir := "internal/database/migrations"
	entries, err := uc.fs.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("could not read %s: %w", migrationDir, err)
	}

	var sqlFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			sqlFiles = append(sqlFiles, filepath.Join(migrationDir, e.Name()))
		}
	}
	if len(sqlFiles) == 0 {
		return fmt.Errorf("no migration files found in %s", migrationDir)
	}
	sort.Strings(sqlFiles)
	latest := sqlFiles[len(sqlFiles)-1]

	send(progress, StepStarted{Label: fmt.Sprintf("Deleting %s", filepath.Base(latest))})
	if err := uc.fs.Remove(latest); err != nil {
		send(progress, StepFailed{Label: filepath.Base(latest), Err: err})
		return err
	}
	send(progress, StepDone{Label: fmt.Sprintf("Deleted %s", filepath.Base(latest))})

	cfg, _ := env.Load()
	atlasEnv := buildAtlasEnv(cfg)

	writer := &progressWriter{progress: progress}
	send(progress, StepStarted{Label: "Updating atlas hash"})
	if err := uc.exec.RunStreamedWithEnv(
		context.Background(), atlasEnv, writer,
		"atlas", "migrate", "hash", "--env", "gorm", "--config", atlasConfigURI,
	); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("atlas hash: %v", err)})
	} else {
		send(progress, StepDone{Label: "Atlas hash updated"})
	}

	send(progress, UseCaseDone{})
	return nil
}
