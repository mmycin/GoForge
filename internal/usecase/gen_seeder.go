package usecase

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mmycin/GoForge/internal/scaffold"
)

// GenSeederUseCase generates a new database seeder file.
type GenSeederUseCase struct {
	gen *scaffold.Generator
}

// NewGenSeederUseCase constructs a GenSeederUseCase.
func NewGenSeederUseCase(gen *scaffold.Generator) *GenSeederUseCase {
	return &GenSeederUseCase{gen: gen}
}

// Run generates internal/database/seeders/<safe>_seeder.go.
func (uc *GenSeederUseCase) Run(name string, progress Progress) error {
	// Normalise: strip any trailing "Seeder" or "seeder" suffix so that both
	// "user" and "UserSeeder" produce the same output.
	base := strings.TrimSuffix(name, "Seeder")
	base = strings.TrimSuffix(base, "seeder")
	base = strings.TrimSuffix(base, "_seeder")

	safeID := scaffold.SafeIdentifier(strings.ToLower(base))
	typeName := scaffold.ToCamelCase(safeID) + "Seeder"
	dest := filepath.Join("internal", "database", "seeders", safeID+"_seeder.go")

	if uc.gen.Exists(dest) {
		return fmt.Errorf("seeder file already exists at %s", dest)
	}

	data := scaffold.SeederData{
		PackageName: safeID,
		TypeName:    typeName,
	}

	send(progress, StepStarted{Label: "Writing " + dest})
	if err := uc.gen.Generate("seeder/seeder.go.tmpl", data, dest); err != nil {
		send(progress, StepFailed{Label: dest, Err: err})
		return err
	}
	send(progress, StepDone{Label: "Writing " + dest})
	send(progress, LogLine{Level: "info", Text: fmt.Sprintf("Next: open %s and add your seeding logic, then register &%s{} in DatabaseSeeder.Run()", dest, typeName)})
	send(progress, UseCaseDone{Files: []GeneratedFile{{Path: dest}}})
	return nil
}
