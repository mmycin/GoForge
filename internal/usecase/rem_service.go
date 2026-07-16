package usecase

import (
	"fmt"
	"path/filepath"

	"github.com/mmycin/GoForge/internal/infra"
	"github.com/mmycin/GoForge/internal/scaffold"
)

// RemServiceUseCase removes a service and its associated proto files.
type RemServiceUseCase struct {
	gen *scaffold.Generator
	fs  infra.FileSystem
}

// NewRemServiceUseCase constructs a RemServiceUseCase.
func NewRemServiceUseCase(gen *scaffold.Generator, fs infra.FileSystem) *RemServiceUseCase {
	return &RemServiceUseCase{gen: gen, fs: fs}
}

// Run removes the service directory, its proto directory, and regenerates kernel.go.
func (uc *RemServiceUseCase) Run(name, modulePath string, progress Progress) error {
	servicesDir := filepath.Join("internal/services", name)
	protoDir := filepath.Join("internal/proto", name)

	if !uc.gen.Exists(servicesDir) {
		return fmt.Errorf("service %q does not exist", name)
	}

	send(progress, StepStarted{Label: fmt.Sprintf("Removing internal/services/%s", name)})
	if err := uc.fs.RemoveAll(servicesDir); err != nil {
		send(progress, StepFailed{Label: "Remove service directory", Err: err})
		return err
	}
	send(progress, StepDone{Label: fmt.Sprintf("Removed internal/services/%s", name)})

	send(progress, StepStarted{Label: fmt.Sprintf("Removing internal/proto/%s", name)})
	if err := uc.fs.RemoveAll(protoDir); err != nil {
		// Non-fatal — proto dir may not exist yet.
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("proto dir: %v", err)})
	} else {
		send(progress, StepDone{Label: fmt.Sprintf("Removed internal/proto/%s", name)})
	}

	send(progress, StepStarted{Label: "Updating kernel.go"})
	if err := regenerateKernel(uc.fs, modulePath); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("kernel.go: %v", err)})
	} else {
		send(progress, StepDone{Label: "Updated kernel.go"})
	}

	send(progress, UseCaseDone{})
	return nil
}
