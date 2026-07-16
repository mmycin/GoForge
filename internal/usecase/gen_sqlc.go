package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mmycin/GoForge/internal/env"
	"github.com/mmycin/GoForge/internal/infra"
)

const sqlcConfig = "core/database/sqlc.yaml"

// GenSqlcUseCase runs sqlc generate and injects SQLC into database.go.
type GenSqlcUseCase struct {
	fs   infra.FileSystem
	exec infra.Executor
}

// NewGenSqlcUseCase constructs a GenSqlcUseCase.
func NewGenSqlcUseCase(fs infra.FileSystem, exec infra.Executor) *GenSqlcUseCase {
	return &GenSqlcUseCase{fs: fs, exec: exec}
}

// SqlcInfo contains the pre-flight info shown to the user.
type SqlcInfo struct {
	Engine string
	Config string
	Output string
}

// Detect reads the current engine from .env.
func (uc *GenSqlcUseCase) Detect() SqlcInfo {
	cfg, _ := env.Load()
	engine := "sqlite"
	if cfg != nil && cfg.DBConnection != "" {
		engine = cfg.DBConnection
	}
	if engine == "postgres" {
		engine = "postgresql"
	}
	return SqlcInfo{Engine: engine, Config: sqlcConfig, Output: "core/database/gen/"}
}

// Run generates SQLC code.
func (uc *GenSqlcUseCase) Run(progress Progress) error {
	info := uc.Detect()

	send(progress, StepStarted{Label: "Updating sqlc.yaml engine"})
	if err := uc.updateSqlcConfig(info.Engine); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("sqlc.yaml: %v", err)})
	} else {
		send(progress, StepDone{Label: "sqlc.yaml updated"})
	}

	send(progress, StepStarted{Label: "Transforming query placeholders"})
	if err := uc.transformQueries(info.Engine); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("transform: %v", err)})
	} else {
		send(progress, StepDone{Label: "Query placeholders transformed"})
	}

	writer := &progressWriter{progress: progress}
	send(progress, StepStarted{Label: "Running sqlc generate"})
	if err := uc.exec.RunStreamed(context.Background(), writer, "sqlc", "generate", "--file", sqlcConfig); err != nil {
		send(progress, StepFailed{Label: "sqlc generate", Err: err})
		return fmt.Errorf("sqlc generate failed: %w", err)
	}
	send(progress, StepDone{Label: "sqlc generate completed"})

	send(progress, StepStarted{Label: "Injecting SQLC into database.go"})
	cfg, _ := env.Load()
	modulePath := "github.com/mmycin/goforge"
	if cfg != nil && cfg.Module != "" {
		modulePath = cfg.Module
	}
	if err := uc.injectSqlc("core/database/database.go", modulePath); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("inject: %v", err)})
	} else {
		send(progress, StepDone{Label: "SQLC injected into database.go"})
	}

	send(progress, UseCaseDone{})
	return nil
}

func (uc *GenSqlcUseCase) updateSqlcConfig(engine string) error {
	content, err := uc.fs.ReadFile(sqlcConfig)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, "engine:") {
			parts := strings.SplitN(line, "engine:", 2)
			if len(parts) == 2 {
				indent := parts[0]
				suffix := ""
				if idx := strings.Index(parts[1], "#"); idx != -1 {
					suffix = " " + parts[1][idx:]
				}
				lines[i] = fmt.Sprintf("%sengine: %q%s", indent, engine, suffix)
			}
			break
		}
	}
	return uc.fs.WriteFile(sqlcConfig, []byte(strings.Join(lines, "\n")), 0644)
}

func (uc *GenSqlcUseCase) transformQueries(engine string) error {
	entries, err := uc.fs.ReadDir("internal/database/queries")
	if err != nil {
		return nil // directory may not exist yet
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		path := filepath.Join("internal/database/queries", e.Name())
		content, err := uc.fs.ReadFile(path)
		if err != nil {
			continue
		}
		s := string(content)
		if engine == "postgresql" {
			lines := strings.Split(s, "\n")
			idx := 1
			for i, line := range lines {
				if strings.Contains(line, "-- name:") {
					idx = 1
				}
				for strings.Contains(lines[i], "?") {
					lines[i] = strings.Replace(lines[i], "?", fmt.Sprintf("$%d", idx), 1)
					idx++
				}
			}
			s = strings.Join(lines, "\n")
		} else {
			for n := 1; n < 50; n++ {
				s = strings.ReplaceAll(s, fmt.Sprintf("$%d", n), "?")
			}
		}
		_ = uc.fs.WriteFile(path, []byte(s), 0644)
	}
	return nil
}

func (uc *GenSqlcUseCase) injectSqlc(targetPath, modulePath string) error {
	content, err := uc.fs.ReadFile(targetPath)
	if err != nil {
		return err
	}
	code := string(content)
	if strings.Contains(code, "sqlc.New(sqlDB)") {
		return nil // already injected
	}

	lines := strings.Split(code, "\n")
	var out []string
	importAdded, fieldAdded, sqlDBAdded, litUpdated := false, false, false, false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if !importAdded && strings.Contains(line, "core/config") {
			out = append(out, line)
			out = append(out, fmt.Sprintf("\tsqlc \"%s/core/database/gen\"", modulePath))
			importAdded = true
			continue
		}
		if !fieldAdded && trimmed == "Gorm *gorm.DB" {
			out = append(out, line, "\tSqlc *sqlc.Queries")
			fieldAdded = true
			continue
		}
		if !sqlDBAdded && trimmed == "if err != nil {" && i > 0 && strings.Contains(lines[i-1], "gorm.Open") {
			out = append(out, line, lines[i+1], lines[i+2])
			i += 2
			out = append(out, "", "\tsqlDB, err := gormDB.DB()", "\tif err != nil {", "\t\treturn err", "\t}")
			sqlDBAdded = true
			continue
		}
		if !litUpdated && trimmed == "Gorm: gormDB," {
			out = append(out, line, "\t\tSqlc: sqlc.New(sqlDB),")
			litUpdated = true
			continue
		}
		out = append(out, line)
	}
	return uc.fs.WriteFile(targetPath, []byte(strings.Join(out, "\n")), 0644)
}
