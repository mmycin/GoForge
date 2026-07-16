package usecase

import (
	"fmt"
	"path/filepath"

	"github.com/mmycin/GoForge/internal/scaffold"
)

// GenCommandUseCase generates a custom console command file.
type GenCommandUseCase struct {
	gen *scaffold.Generator
}

// NewGenCommandUseCase constructs a GenCommandUseCase.
func NewGenCommandUseCase(gen *scaffold.Generator) *GenCommandUseCase {
	return &GenCommandUseCase{gen: gen}
}

// Run generates internal/console/<safe>_cmd.go.
func (uc *GenCommandUseCase) Run(name string, progress Progress) error {
	safeID := scaffold.SafeIdentifier(name)
	typeName := scaffold.ToCamelCase(safeID) + "Cmd"
	dest := filepath.Join("internal", "console", safeID+"_cmd.go")

	if uc.gen.Exists(dest) {
		return fmt.Errorf("command file already exists at %s", dest)
	}

	data := scaffold.CommandData{
		PackageName: safeID,
		CommandUse:  name,
		TypeName:    typeName,
	}

	send(progress, StepStarted{Label: "Writing " + dest})
	if err := uc.gen.Generate("command/cmd.go.tmpl", data, dest); err != nil {
		send(progress, StepFailed{Label: dest, Err: err})
		return err
	}
	send(progress, StepDone{Label: "Written " + dest})
	send(progress, UseCaseDone{Files: []GeneratedFile{{Path: dest}}})
	return nil
}
