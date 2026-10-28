package usecase

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mmycin/GoForge/internal/infra"
	"github.com/mmycin/GoForge/internal/scaffold"
)

// RemSeederUseCase deletes an existing database seeder file.
type RemSeederUseCase struct {
	fs infra.FileSystem
}

// NewRemSeederUseCase constructs a RemSeederUseCase.
func NewRemSeederUseCase(fs infra.FileSystem) *RemSeederUseCase {
	return &RemSeederUseCase{fs: fs}
}

// Run removes internal/database/seeders/<safe>_seeder.go.
func (uc *RemSeederUseCase) Run(name string, progress Progress) error {
	base := strings.TrimSuffix(name, "Seeder")
	base = strings.TrimSuffix(base, "seeder")
	base = strings.TrimSuffix(base, "_seeder")

	safeID := scaffold.SafeIdentifier(strings.ToLower(base))
	target := filepath.Join("internal", "database", "seeders", safeID+"_seeder.go")

	if _, err := uc.fs.Stat(target); err != nil {
		return fmt.Errorf("seeder file %q does not exist at %s", name, target)
	}

	send(progress, StepStarted{Label: "Deleting " + target})
	if err := uc.fs.Remove(target); err != nil {
		send(progress, StepFailed{Label: target, Err: err})
		return err
	}
	send(progress, StepDone{Label: "Deleted " + target})
	send(progress, LogLine{Level: "info", Text: fmt.Sprintf("Remember to remove &%sSeeder{} from DatabaseSeeder.Run() in database_seeder.go", scaffold.ToCamelCase(safeID))})
	send(progress, UseCaseDone{})
	return nil
}
