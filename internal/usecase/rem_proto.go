package usecase

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mmycin/GoForge/internal/infra"
)

// RemProtoUseCase removes generated .pb.go files and gen/ directories.
type RemProtoUseCase struct {
	fs infra.FileSystem
}

// NewRemProtoUseCase constructs a RemProtoUseCase.
func NewRemProtoUseCase(fs infra.FileSystem) *RemProtoUseCase {
	return &RemProtoUseCase{fs: fs}
}

// Run removes all generated proto artifacts from internal/services and internal/proto.
func (uc *RemProtoUseCase) Run(progress Progress) error {
	entries, err := uc.fs.ReadDir("internal/services")
	if err != nil {
		return fmt.Errorf("could not read internal/services: %w", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		svcName := e.Name()

		// Remove *.pb.go files inside the service directory.
		svcDir := filepath.Join("internal/services", svcName)
		files, _ := uc.fs.ReadDir(svcDir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".pb.go") {
				path := filepath.Join(svcDir, f.Name())
				send(progress, StepStarted{Label: "Removing " + path})
				if err := uc.fs.Remove(path); err != nil {
					send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("remove %s: %v", path, err)})
				} else {
					send(progress, StepDone{Label: "Removed " + path})
				}
			}
		}

		// Remove internal/proto/<svcName>/gen/
		genDir := filepath.Join("internal/proto", svcName, "gen")
		if _, statErr := uc.fs.Stat(genDir); statErr == nil {
			send(progress, StepStarted{Label: "Removing " + genDir})
			if err := uc.fs.RemoveAll(genDir); err != nil {
				send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("remove %s: %v", genDir, err)})
			} else {
				send(progress, StepDone{Label: "Removed " + genDir})
			}
		}
	}

	send(progress, UseCaseDone{})
	return nil
}
