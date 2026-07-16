package usecase

import (
	"fmt"
	"path/filepath"

	"github.com/mmycin/GoForge/internal/infra"
	"github.com/mmycin/GoForge/internal/scaffold"
)

// RemCommandUseCase deletes a custom console command file.
type RemCommandUseCase struct {
	fs infra.FileSystem
}

// NewRemCommandUseCase constructs a RemCommandUseCase.
func NewRemCommandUseCase(fs infra.FileSystem) *RemCommandUseCase {
	return &RemCommandUseCase{fs: fs}
}

// Run removes core/console/<safe>_cmd.go.
func (uc *RemCommandUseCase) Run(name string, progress Progress) error {
	safeID := scaffold.SafeIdentifier(name)
	target := filepath.Join("core", "console", safeID+"_cmd.go")

	if _, err := uc.fs.Stat(target); err != nil {
		return fmt.Errorf("command file %q does not exist at %s", name, target)
	}

	send(progress, StepStarted{Label: "Deleting " + target})
	if err := uc.fs.Remove(target); err != nil {
		send(progress, StepFailed{Label: target, Err: err})
		return err
	}
	send(progress, StepDone{Label: "Deleted " + target})
	send(progress, UseCaseDone{})
	return nil
}
