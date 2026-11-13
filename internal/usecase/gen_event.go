package usecase

import (
	"fmt"
	"path/filepath"

	"github.com/mmycin/GoForge/internal/infra"
	"github.com/mmycin/GoForge/internal/scaffold"
)

// GenEventUseCase generates events.go and listeners.go for a service and
// registers RegisterListeners in the service's module.go Invokes slice.
type GenEventUseCase struct {
	gen *scaffold.Generator
	fs  infra.FileSystem
}

// NewGenEventUseCase constructs a GenEventUseCase.
func NewGenEventUseCase(gen *scaffold.Generator, fs infra.FileSystem) *GenEventUseCase {
	return &GenEventUseCase{gen: gen, fs: fs}
}

// Run generates both event files into the given service's directory and wires module.go.
func (uc *GenEventUseCase) Run(serviceName, modulePath string, progress Progress) error {
	serviceDir := filepath.Join("internal/services", serviceName)

	if !uc.gen.Exists(serviceDir) {
		return fmt.Errorf("service %q does not exist — run goforge gen:service %s first", serviceName, serviceName)
	}

	typeName := scaffold.ToCamelCase(serviceName)
	data := scaffold.EventData{
		PackageName: serviceName,
		TypeName:    typeName,
		ModulePath:  modulePath,
	}

	files := []struct {
		tmpl string
		dest string
	}{
		{"event/events.go.tmpl", filepath.Join(serviceDir, "events.go")},
		{"event/listeners.go.tmpl", filepath.Join(serviceDir, "listeners.go")},
	}

	var generated []GeneratedFile
	for _, f := range files {
		if uc.gen.Exists(f.dest) {
			return fmt.Errorf("%s already exists", f.dest)
		}
		send(progress, StepStarted{Label: "Writing " + filepath.Base(f.dest)})
		if err := uc.gen.Generate(f.tmpl, data, f.dest); err != nil {
			send(progress, StepFailed{Label: filepath.Base(f.dest), Err: err})
			return err
		}
		generated = append(generated, GeneratedFile{Path: f.dest})
		send(progress, StepDone{Label: "Writing " + filepath.Base(f.dest)})
	}

	moduleFile := filepath.Join(serviceDir, "module.go")
	if uc.gen.Exists(moduleFile) {
		send(progress, StepStarted{Label: "Updating module.go"})
		updated, err := addEntryToModuleSlice(uc.fs, moduleFile, "Invokes", "RegisterListeners")
		if err != nil {
			send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("module.go update: %v", err)})
		} else {
			if updated {
				generated = append(generated, GeneratedFile{Path: moduleFile, Updated: true})
			}
			send(progress, StepDone{Label: "Updating module.go"})
		}
	}

	send(progress, UseCaseDone{Files: generated})
	return nil
}
