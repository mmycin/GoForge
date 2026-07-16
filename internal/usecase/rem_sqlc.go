package usecase

import (
	"fmt"
	"strings"

	"github.com/mmycin/GoForge/internal/infra"
)

// RemSqlcUseCase removes SQLC integration from database.go and deletes core/database/gen/.
type RemSqlcUseCase struct {
	fs infra.FileSystem
}

// NewRemSqlcUseCase constructs a RemSqlcUseCase.
func NewRemSqlcUseCase(fs infra.FileSystem) *RemSqlcUseCase {
	return &RemSqlcUseCase{fs: fs}
}

// Run removes SQLC from database.go and deletes the generated output directory.
func (uc *RemSqlcUseCase) Run(progress Progress) error {
	targetPath := "core/database/database.go"

	send(progress, StepStarted{Label: "Reverting database.go"})
	if err := uc.revertDatabaseGo(targetPath); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("revert: %v", err)})
	} else {
		send(progress, StepDone{Label: "database.go reverted"})
	}

	genDir := "core/database/gen"
	if _, err := uc.fs.Stat(genDir); err == nil {
		send(progress, StepStarted{Label: "Deleting core/database/gen/"})
		if err := uc.fs.RemoveAll(genDir); err != nil {
			send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("remove gen: %v", err)})
		} else {
			send(progress, StepDone{Label: "Deleted core/database/gen/"})
		}
	}

	send(progress, UseCaseDone{})
	return nil
}

func (uc *RemSqlcUseCase) revertDatabaseGo(path string) error {
	content, err := uc.fs.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if strings.Contains(line, "core/database/gen") {
			continue
		}
		if trimmed == "Sqlc *sqlc.Queries" {
			out = append(out, "\t// Sqlc field removed — run goforge gen:sqlc to restore")
			continue
		}
		if trimmed == "sqlDB, err := gormDB.DB()" {
			i += 3
			if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == "" {
				i++
			}
			continue
		}
		if strings.Contains(line, "Sqlc: sqlc.New(sqlDB),") {
			continue
		}
		out = append(out, line)
	}
	return uc.fs.WriteFile(path, []byte(strings.Join(out, "\n")), 0644)
}
