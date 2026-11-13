package usecase

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/mmycin/GoForge/internal/infra"
	"github.com/mmycin/GoForge/internal/scaffold"
)

// GenServiceUseCase generates the full service layer for a given name.
type GenServiceUseCase struct {
	gen *scaffold.Generator
	fs  infra.FileSystem
}

// NewGenServiceUseCase constructs a GenServiceUseCase.
func NewGenServiceUseCase(gen *scaffold.Generator, fs infra.FileSystem) *GenServiceUseCase {
	return &GenServiceUseCase{gen: gen, fs: fs}
}

// Run generates all service files (including module.go) and updates kernel.go.
func (uc *GenServiceUseCase) Run(name, modulePath string, progress Progress) error {
	targetDir := filepath.Join("internal/services", name)

	if uc.gen.Exists(targetDir) {
		return fmt.Errorf("service %q already exists at %s", name, targetDir)
	}

	typeName := scaffold.ToCamelCase(name)
	data := scaffold.ServiceData{
		PackageName: name,
		TypeName:    typeName,
		ModulePath:  modulePath,
	}

	serviceFiles := []struct {
		tmpl string
		dest string
	}{
		{"service/service.go.tmpl", filepath.Join(targetDir, "service.go")},
		{"service/handler.go.tmpl", filepath.Join(targetDir, "handler.go")},
		{"service/model.go.tmpl", filepath.Join(targetDir, "model.go")},
		{"service/routes.go.tmpl", filepath.Join(targetDir, "routes.go")},
		{"service/docs.go.tmpl", filepath.Join(targetDir, "docs.go")},
		{"service/grpc.go.tmpl", filepath.Join(targetDir, "grpc.go")},
		{"service/module.go.tmpl", filepath.Join(targetDir, "module.go")},
	}

	var generated []GeneratedFile
	for _, f := range serviceFiles {
		send(progress, StepStarted{Label: "Writing " + filepath.Base(f.dest)})
		if err := uc.gen.Generate(f.tmpl, data, f.dest); err != nil {
			send(progress, StepFailed{Label: filepath.Base(f.dest), Err: err})
			return err
		}
		generated = append(generated, GeneratedFile{Path: f.dest})
		send(progress, StepDone{Label: "Writing " + filepath.Base(f.dest)})
	}

	// Proto file
	protoDir := filepath.Join("internal/proto", name)
	protoDest := filepath.Join(protoDir, name+".proto")
	protoData := scaffold.ProtoData{PackageName: name, TypeName: typeName, ModulePath: modulePath}
	send(progress, StepStarted{Label: "Writing proto file"})
	if err := uc.gen.Generate("proto/service.proto.tmpl", protoData, protoDest); err != nil {
		send(progress, StepFailed{Label: "Writing proto file", Err: err})
		// non-fatal
	} else {
		generated = append(generated, GeneratedFile{Path: protoDest})
		send(progress, StepDone{Label: "Writing proto file"})
	}

	// Regenerate kernel.go
	send(progress, StepStarted{Label: "Updating kernel.go"})
	if err := uc.regenerateKernel(modulePath); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("kernel.go update: %v", err)})
	} else {
		generated = append(generated, GeneratedFile{Path: "internal/services/kernel.go", Updated: true})
		send(progress, StepDone{Label: "Updating kernel.go"})
	}

	send(progress, UseCaseDone{Files: generated})
	return nil
}

// ─── kernel.go regeneration ──────────────────────────────────────────────────

func (uc *GenServiceUseCase) regenerateKernel(modulePath string) error {
	return regenerateKernel(uc.fs, modulePath)
}

// regenerateKernel scans internal/services/* for directories containing module.go
// and regenerates internal/services/kernel.go with a minimal Modules() registry
// and Model() aggregator.
func regenerateKernel(fsys infra.FileSystem, modulePath string) error {
	entries, err := fsys.ReadDir("internal/services")
	if err != nil {
		return err
	}

	var serviceNames []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		moduleFile := filepath.Join("internal/services", e.Name(), "module.go")
		if _, statErr := fsys.Stat(moduleFile); statErr != nil {
			continue
		}
		serviceNames = append(serviceNames, e.Name())
	}
	sort.Strings(serviceNames)

	// Remove legacy kernel_models.go if present from older projects.
	if _, statErr := fsys.Stat("internal/services/kernel_models.go"); statErr == nil {
		_ = fsys.Remove("internal/services/kernel_models.go")
	}

	const kernelTmpl = `package services

import (
	"{{.ModulePath}}/core/di"
{{- range .Services}}
	"{{$.ModulePath}}/internal/services/{{.}}"
{{- end}}
)

// Modules returns all application modules registered in the central service kernel.
func Modules() []di.Module {
	return []di.Module{
{{- range .Services}}
		{{.}}.Module(),
{{- end}}
	}
}

// Model aggregates all GORM models from registered modules for migrations.
func Model() []any {
	return di.CollectModels(Modules()...)
}
`

	t, err := template.New("kernel").Parse(kernelTmpl)
	if err != nil {
		return err
	}

	var buf strings.Builder
	data := struct {
		ModulePath string
		Services   []string
	}{
		ModulePath: modulePath,
		Services:   serviceNames,
	}
	if err := t.Execute(&buf, data); err != nil {
		return err
	}

	return fsys.WriteFile("internal/services/kernel.go", []byte(buf.String()), 0644)
}
