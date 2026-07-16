// Package cmd wires the full dependency graph once at startup and passes it
// into every cobra command via closure.  Nothing in cmd/ calls os directly —
// all I/O goes through the infra interfaces.
package cmd

import (
	"github.com/mmycin/GoForge/internal/infra"
	"github.com/mmycin/GoForge/internal/scaffold"
	"github.com/mmycin/GoForge/internal/usecase"
)

// Deps is the assembled dependency graph shared by all cobra Run functions.
type Deps struct {
	FS        infra.FileSystem
	Exec      infra.Executor
	Git       infra.Git
	Generator *scaffold.Generator

	NewProject   *usecase.NewProjectUseCase
	GenService   *usecase.GenServiceUseCase
	RemService   *usecase.RemServiceUseCase
	GenEvent     *usecase.GenEventUseCase
	GenCommand   *usecase.GenCommandUseCase
	RemCommand   *usecase.RemCommandUseCase
	GenConfig    *usecase.GenConfigUseCase
	GenProto     *usecase.GenProtoUseCase
	RemProto     *usecase.RemProtoUseCase
	Migrate      *usecase.MigrateUseCase
	GenMigration *usecase.GenMigrationUseCase
	RemMigration *usecase.RemMigrationUseCase
	GenSqlc      *usecase.GenSqlcUseCase
	RemSqlc      *usecase.RemSqlcUseCase
}

// NewDeps constructs the full dependency graph with real OS implementations.
// Called once from root.go before cobra parses subcommands.
func NewDeps() *Deps {
	fs := infra.NewOsFileSystem()
	exec := infra.NewOsExecutor()
	git := infra.NewOsGit(exec)
	gen := scaffold.NewGenerator(fs)

	return &Deps{
		FS:        fs,
		Exec:      exec,
		Git:       git,
		Generator: gen,

		NewProject:   usecase.NewNewProjectUseCase(fs, exec, git),
		GenService:   usecase.NewGenServiceUseCase(gen, fs),
		RemService:   usecase.NewRemServiceUseCase(gen, fs),
		GenEvent:     usecase.NewGenEventUseCase(gen),
		GenCommand:   usecase.NewGenCommandUseCase(gen),
		RemCommand:   usecase.NewRemCommandUseCase(fs),
		GenConfig:    usecase.NewGenConfigUseCase(gen, fs),
		GenProto:     usecase.NewGenProtoUseCase(gen, fs, exec),
		RemProto:     usecase.NewRemProtoUseCase(fs),
		Migrate:      usecase.NewMigrateUseCase(fs, exec),
		GenMigration: usecase.NewGenMigrationUseCase(fs, exec),
		RemMigration: usecase.NewRemMigrationUseCase(fs, exec),
		GenSqlc:      usecase.NewGenSqlcUseCase(fs, exec),
		RemSqlc:      usecase.NewRemSqlcUseCase(fs),
	}
}
