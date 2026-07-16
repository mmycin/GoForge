package usecase

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mmycin/GoForge/internal/infra"
	"github.com/mmycin/GoForge/internal/scaffold"
)

// GenConfigUseCase scaffolds a new config section and wires it into AllConfig.
type GenConfigUseCase struct {
	gen *scaffold.Generator
	fs  infra.FileSystem
}

// NewGenConfigUseCase constructs a GenConfigUseCase.
func NewGenConfigUseCase(gen *scaffold.Generator, fs infra.FileSystem) *GenConfigUseCase {
	return &GenConfigUseCase{gen: gen, fs: fs}
}

// Run generates core/config/<name>.go and patches core/config/config.go.
func (uc *GenConfigUseCase) Run(name, modulePath string, progress Progress) error {
	configDir := "core/config"

	if _, err := uc.fs.Stat(configDir); err != nil {
		return fmt.Errorf("core/config/ does not exist — are you in the project root?")
	}

	camel := scaffold.ToCamelCase(name)
	targetFile := filepath.Join(configDir, name+".go")

	if uc.gen.Exists(targetFile) {
		return fmt.Errorf("core/config/%s.go already exists", name)
	}

	data := scaffold.ConfigData{
		PackageName: name,
		TypeName:    camel,
		EnvPrefix:   scaffold.EnvPrefix(name),
		ModulePath:  modulePath,
	}

	send(progress, StepStarted{Label: fmt.Sprintf("Writing core/config/%s.go", name)})
	if err := uc.gen.Generate("config/config.go.tmpl", data, targetFile); err != nil {
		send(progress, StepFailed{Label: targetFile, Err: err})
		return err
	}
	send(progress, StepDone{Label: fmt.Sprintf("Written core/config/%s.go", name)})

	send(progress, StepStarted{Label: "Wiring into core/config/config.go"})
	if err := uc.wireIntoAllConfig(name, camel); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("auto-wire failed: %v", err)})
	} else {
		send(progress, StepDone{Label: "Wired into core/config/config.go"})
	}

	send(progress, UseCaseDone{Files: []GeneratedFile{
		{Path: targetFile},
		{Path: "core/config/config.go", Updated: true},
	}})
	return nil
}

func (uc *GenConfigUseCase) wireIntoAllConfig(name, camel string) error {
	configPath := filepath.Join("core", "config", "config.go")
	raw, err := uc.fs.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("could not read config.go: %w", err)
	}
	src := string(raw)

	fieldLine := "\t" + fmt.Sprintf("%-7s %sConfig", camel, camel)
	loadBlock := fmt.Sprintf("\tif cfg.%s, err = load%sConfig(); err != nil {\n\t\treturn nil, err\n\t}", camel, camel)
	assignLine := fmt.Sprintf("\t%s = cfg.%s", camel, camel)

	if !strings.Contains(src, fieldLine) {
		src = insertBefore(src, "type AllConfig struct {", "\n}", "\n"+fieldLine)
	}
	if !strings.Contains(src, "load"+camel+"Config()") {
		anchor := "\t// Set package-level variables"
		if !strings.Contains(src, anchor) {
			anchor = "\tApp = cfg.App"
		}
		if strings.Contains(src, anchor) {
			src = strings.Replace(src, anchor, loadBlock+"\n\n\t"+strings.TrimSpace(anchor), 1)
		}
	}
	if !strings.Contains(src, assignLine) {
		src = strings.Replace(src, "\treturn &cfg, nil", assignLine+"\n\treturn &cfg, nil", 1)
	}

	return uc.fs.WriteFile(configPath, []byte(src), 0644)
}

func insertBefore(src, scopeMarker, insertBefore, content string) string {
	si := strings.Index(src, scopeMarker)
	if si == -1 {
		return src
	}
	after := src[si:]
	ri := strings.Index(after, insertBefore)
	if ri == -1 {
		return src
	}
	abs := si + ri
	return src[:abs] + content + src[abs:]
}
