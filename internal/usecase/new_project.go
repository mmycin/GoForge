package usecase

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/mmycin/GoForge/internal/infra"
)

const templateRepo = "https://github.com/mmycin/goforge-template"
const templateBranch = "main"
const oldModulePlaceholder = "github.com/mmycin/goforge"

// NewProjectInput holds all data collected by the wizard.
type NewProjectInput struct {
	ProjectName string
	ModulePath  string
	Databases   []string // e.g. ["sqlite", "postgresql"]
	InitGit     bool
}

// driverFiles maps a driver key to its file path inside the cloned project.
var driverFiles = map[string]string{
	"sqlite":     "core/database/drivers/sqlite.go",
	"mysql":      "core/database/drivers/mysql.go",
	"postgresql": "core/database/drivers/postgresql.go",
	"sqlserver":  "core/database/drivers/sqlserver.go",
}

// NewProjectUseCase creates a new GoForge project from the official template.
type NewProjectUseCase struct {
	fs   infra.FileSystem
	exec infra.Executor
	git  infra.Git
}

// NewNewProjectUseCase constructs a NewProjectUseCase with the provided dependencies.
func NewNewProjectUseCase(fs infra.FileSystem, exec infra.Executor, git infra.Git) *NewProjectUseCase {
	return &NewProjectUseCase{fs: fs, exec: exec, git: git}
}

// Run executes the full project creation pipeline, pushing events onto progress.
// It is designed to be called from a goroutine so the TUI can remain responsive.
func (uc *NewProjectUseCase) Run(input NewProjectInput, progress Progress) error {
	project := input.ProjectName

	// ── Step 1: Clone ────────────────────────────────────────────────────────
	send(progress, StepStarted{Label: "Creating project structure"})
	if err := uc.git.Clone(templateRepo, project, templateBranch); err != nil {
		send(progress, StepFailed{Label: "Creating project structure", Err: err})
		return fmt.Errorf("clone failed: %w", err)
	}
	// Remove the template's .git history so the project starts clean.
	_ = uc.fs.RemoveAll(filepath.Join(project, ".git"))
	send(progress, StepDone{Label: "Creating project structure"})

	// ── Step 2: Replace module path ─────────────────────────────────────────
	send(progress, StepStarted{Label: "Initialising module path"})
	if err := uc.replaceModule(project, input.ModulePath); err != nil {
		send(progress, StepFailed{Label: "Initialising module path", Err: err})
		return err
	}
	send(progress, StepDone{Label: "Initialising module path"})

	// ── Step 3: Remove unused database drivers ───────────────────────────────
	send(progress, StepStarted{Label: "Initializing database drivers"})
	if err := uc.removeDrivers(project, input.Databases, progress); err != nil {
		send(progress, StepFailed{Label: "Initializing database drivers", Err: err})
		return err
	}
	send(progress, StepDone{Label: "Initializing database drivers"})

	// ── Step 4: go mod tidy — must run inside the project dir ───────────────
	send(progress, StepStarted{Label: "Installing dependencies"})
	if err := uc.exec.RunInDir(project, "go", "mod", "tidy"); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("go mod tidy: %v", err)})
	}
	send(progress, StepDone{Label: "Installing dependencies"})

	// ── Step 5 & 6: Git init (optional) ─────────────────────────────────────
	if input.InitGit && uc.git.IsAvailable() {
		send(progress, StepStarted{Label: "Initialising git repository"})
		if err := uc.git.Init(project); err != nil {
			send(progress, StepFailed{Label: "Initialising git repository", Err: err})
		} else {
			send(progress, StepDone{Label: "Initialising git repository"})

			send(progress, StepStarted{Label: "Creating initial commit"})
			_ = uc.git.AddAll(project)
			if err := uc.git.Commit(project, "chore: initial commit from GoForge"); err != nil {
				send(progress, StepFailed{Label: "Creating initial commit", Err: err})
			} else {
				send(progress, StepDone{Label: "Creating initial commit"})
			}
		}
	}

	send(progress, UseCaseDone{})
	return nil
}

func (uc *NewProjectUseCase) replaceModule(projectDir, newModule string) error {
	return uc.fs.WalkDir(projectDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		content, err := uc.fs.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(content), oldModulePlaceholder) {
			replaced := strings.ReplaceAll(string(content), oldModulePlaceholder, newModule)
			return uc.fs.WriteFile(path, []byte(replaced), 0644)
		}
		return nil
	})
}

func (uc *NewProjectUseCase) removeDrivers(projectDir string, selected []string, progress Progress) error {
	selectedSet := make(map[string]bool, len(selected))
	for _, k := range selected {
		selectedSet[k] = true
	}
	for key, rel := range driverFiles {
		if selectedSet[key] {
			continue
		}
		full := filepath.Join(projectDir, rel)
		if err := uc.fs.Remove(full); err != nil {
			send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("could not remove %s: %v", rel, err)})
		} else {
			send(progress, LogLine{Level: "info", Text: fmt.Sprintf("removed driver: %s", key)})
		}
	}
	return nil
}
