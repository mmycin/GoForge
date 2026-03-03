package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mmycin/GoForge/internal/env"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(genMigrationCmd)
	rootCmd.AddCommand(remMigrationCmd)
	rootCmd.AddCommand(loaderCmd)
	rootCmd.AddCommand(genSqlcCmd)
	rootCmd.AddCommand(remSqlcCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Execute migrations to synchronize database schema with models.`,
	Run: func(cmd *cobra.Command, args []string) {
		Info("Running database migration...")
		migrateDB()
	},
}

var genMigrationCmd = &cobra.Command{
	Use:   "gen:migration [name]",
	Short: "Create a new database migration",
	Long:  `Generate a new database migration file with the specified name.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		Info("Creating migration: %s", name)
		genMigration(name)
	},
}

var remMigrationCmd = &cobra.Command{
	Use:   "rem:migration",
	Short: "Remove the latest database migration",
	Long:  `Delete the most recent migration file and revert the atlas hash.`,
	Run: func(cmd *cobra.Command, args []string) {
		Info("Removing latest migration...")
		remMigration()
	},
}

var loaderCmd = &cobra.Command{
	Use:   "loader",
	Short: "Run GORM schema loader",
	Long:  `Load and display GORM schema definitions.`,
	Run: func(cmd *cobra.Command, args []string) {
		runLoader()
	},
}

var genSqlcCmd = &cobra.Command{
	Use:   "gen:sqlc",
	Short: "Run SQLC code generation",
	Long:  `Execute sqlc generate to create database query code.`,
	Run: func(cmd *cobra.Command, args []string) {
		Info("Running code generation...")
		genSqlc()
	},
}

var remSqlcCmd = &cobra.Command{
	Use:   "rem:sqlc",
	Short: "Remove SQLC integration",
	Long:  `Remove generated SQLC code and revert database kernel integration.`,
	Run: func(cmd *cobra.Command, args []string) {
		Info("Removing SQLC integration...")
		removeSqlc("internal/database/database.go")
	},
}

func genSqlc() {
	if err := updateSqlcConfig(); err != nil {
		Warning("Failed to update sqlc.yaml: %v", err)
	}

	cfg, _ := env.Load()
	engine := "sqlite"
	if cfg != nil && cfg.DBConnection != "" {
		engine = cfg.DBConnection
	}

	if engine == "postgres" || engine == "postgresql" {
		engine = "postgresql"
	} else if engine == "mysql" {
		engine = "mysql"
	} else {
		engine = "sqlite"
	}

	Info("Transforming queries for %s engine...", engine)
	if err := transformQueries(engine); err != nil {
		Warning("Failed to transform queries: %v", err)
	}

	Info("Executing sqlc generate...")
	cmd := exec.Command("sqlc", "generate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		ErrorLog("sqlc generate failed: %v", err)
		os.Exit(1)
	}
	Success("Code generation completed successfully")

	injectSqlc("internal/database/database.go")
}

func transformQueries(engine string) error {
	queriesDir := "internal/database/queries"
	files, err := filepath.Glob(filepath.Join(queriesDir, "*.sql"))
	if err != nil {
		return err
	}

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return err
		}

		newContent := string(content)
		if engine == "postgresql" {
			lines := strings.Split(newContent, "\n")
			placeholderIdx := 1
			for i, line := range lines {
				if strings.Contains(line, "-- name:") {
					placeholderIdx = 1
				}
				for strings.Contains(lines[i], "?") {
					lines[i] = strings.Replace(lines[i], "?", fmt.Sprintf("$%d", placeholderIdx), 1)
					placeholderIdx++
				}
			}
			newContent = strings.Join(lines, "\n")
		} else if engine == "mysql" || engine == "sqlite" {
			for i := 1; i < 50; i++ {
				newContent = strings.ReplaceAll(newContent, fmt.Sprintf("$%d", i), "?")
			}
		}

		if err := os.WriteFile(f, []byte(newContent), 0644); err != nil {
			return err
		}
	}
	return nil
}

func updateSqlcConfig() error {
	configPath := "sqlc.yaml"
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	cfg, _ := env.Load()
	engine := "sqlite"
	if cfg != nil && cfg.DBConnection != "" {
		engine = cfg.DBConnection
	}

	if engine == "postgres" {
		engine = "postgresql"
	}

	lines := strings.Split(string(content), "\n")
	updated := false
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
				updated = true
				break
			}
		}
	}

	if !updated {
		return fmt.Errorf("could not find 'engine' field in %s", configPath)
	}

	return os.WriteFile(configPath, []byte(strings.Join(lines, "\n")), 0644)
}

func injectSqlc(targetPath string) {
	content, err := os.ReadFile(targetPath)
	if err != nil {
		Warning("Could not read %s for injection: %v", targetPath, err)
		return
	}

	code := string(content)
	if strings.Contains(code, "sqlc.New(sqlDB)") {
		return
	}

	Info("Injecting SQLC support into database...")

	cfg, _ := env.Load()
	moduleName := "github.com/mmycin/goforge"
	if cfg != nil && cfg.Module != "" {
		moduleName = cfg.Module
	}

	lines := strings.Split(code, "\n")
	var newLines []string

	importAdded := false
	fieldAdded := false
	sqlDBAdded := false
	literalUpdated := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if !importAdded && strings.Contains(line, "internal/config") {
			newLines = append(newLines, line)
			newLines = append(newLines, fmt.Sprintf("\tsqlc \"%s/internal/database/gen\"", moduleName))
			importAdded = true
			continue
		}

		if !fieldAdded && trimmed == "Gorm *gorm.DB" {
			newLines = append(newLines, line)
			newLines = append(newLines, "\tSqlc *sqlc.Queries")
			fieldAdded = true
			continue
		}

		if !sqlDBAdded && trimmed == "if err != nil {" && i > 0 && strings.Contains(lines[i-1], "gorm.Open") {
			newLines = append(newLines, line)
			newLines = append(newLines, lines[i+1])
			newLines = append(newLines, lines[i+2])
			i += 2

			newLines = append(newLines, "")
			newLines = append(newLines, "\tsqlDB, err := gormDB.DB()")
			newLines = append(newLines, "\tif err != nil {")
			newLines = append(newLines, "\t\treturn err")
			newLines = append(newLines, "\t}")
			sqlDBAdded = true
			continue
		}

		if !literalUpdated && trimmed == "Gorm: gormDB," {
			newLines = append(newLines, line)
			newLines = append(newLines, "\t\tSqlc: sqlc.New(sqlDB),")
			literalUpdated = true
			continue
		}

		newLines = append(newLines, line)
	}

	code = strings.Join(newLines, "\n")
	if err := os.WriteFile(targetPath, []byte(code), 0644); err != nil {
		Warning("Failed to inject SQLC support: %v", err)
	}
	Success("SQLC support injected into database")
}

func removeSqlc(targetPath string) {
	content, err := os.ReadFile(targetPath)
	if err != nil {
		Warning("Could not read %s for removal: %v", targetPath, err)
		return
	}

	code := string(content)
	lines := strings.Split(code, "\n")
	var newLines []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.Contains(line, "internal/database/gen") {
			continue
		}

		if trimmed == "Sqlc *sqlc.Queries" {
			newLines = append(newLines, "\t// Sqlc field will be added when generated code is available")
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

		newLines = append(newLines, line)
	}

	code = strings.Join(newLines, "\n")
	if err := os.WriteFile(targetPath, []byte(code), 0644); err != nil {
		Warning("Failed to remove SQLC support: %v", err)
	}

	genDir := "internal/database/gen"
	if _, err := os.Stat(genDir); err == nil {
		Info("Deleting generated folder: %s", genDir)
		os.RemoveAll(genDir)
	}

	Success("SQLC support removed from database")
}

func remMigration() {
	migrationDir := "internal/database/migrations"
	files, err := filepath.Glob(filepath.Join(migrationDir, "*.sql"))
	if err != nil || len(files) == 0 {
		Info("No migration files found to remove.")
		return
	}

	latest := files[len(files)-1]
	Info("Deleting migration file: %s", latest)
	if err := os.Remove(latest); err != nil {
		ErrorLog("Failed to delete migration file: %v", err)
		return
	}

	Info("Updating atlas migrate hash...")

	cfg, err := env.Load()
	if err != nil {
		Warning("Could not parse env for atlas hash update: %v", err)
	}
	dbConn := "sqlite"
	if cfg != nil && cfg.DBConnection != "" {
		dbConn = cfg.DBConnection
	}

	atlasEnv := os.Environ()
	atlasEnv = append(atlasEnv, "DB_CONNECTION="+dbConn)
	cmd := exec.Command("atlas", "migrate", "hash", "--env", "gorm")
	cmd.Env = atlasEnv
	if err := cmd.Run(); err != nil {
		Warning("Atlas hash update failed: %v", err)
	}

	Success("Latest migration removed successfully")
}

func genMigration(name string) {
	cfg, _ := env.Load()
	dbConn := "sqlite"
	dbName := ""
	dbUser := ""
	dbPass := ""
	dbHost := ""
	dbPort := "3306"
	dbDevName := ""

	if cfg != nil {
		if cfg.DBConnection != "" {
			dbConn = cfg.DBConnection
		}
		if cfg.DBName != "" {
			dbName = cfg.DBName
		}
		if cfg.DBUsername != "" {
			dbUser = cfg.DBUsername
		}
		if cfg.DBPassword != "" {
			dbPass = cfg.DBPassword
		}
		if cfg.DBHost != "" {
			dbHost = cfg.DBHost
		}
		if cfg.DBPort != "" {
			dbPort = cfg.DBPort
		}
		if cfg.DBDevName != "" {
			dbDevName = cfg.DBDevName
		}
	}

	atlasEnv := os.Environ()
	atlasEnv = append(atlasEnv, "DB_CONNECTION="+dbConn)
	atlasEnv = append(atlasEnv, "DB_NAME="+dbName)
	atlasEnv = append(atlasEnv, "DB_USERNAME="+dbUser)
	atlasEnv = append(atlasEnv, "DB_PASSWORD="+dbPass)
	atlasEnv = append(atlasEnv, "DB_HOST="+dbHost)
	atlasEnv = append(atlasEnv, "DB_PORT="+dbPort)
	atlasEnv = append(atlasEnv, "DB_DEV_NAME="+dbDevName)

	Info("Running atlas migrate diff...")
	cmd := exec.Command("atlas", "migrate", "diff", "--env", "gorm", name)
	cmd.Env = atlasEnv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\n")
		ErrorLog("Atlas migration failed: %v", err)
		if dbDevName == "" && (dbConn == "mysql" || dbConn == "postgres") {
			fmt.Println("\nTIP: Atlas requires a clean/empty database for the 'dev' environment.")
			fmt.Printf("1. Create an empty database in your %s server (e.g., 'CREATE DATABASE %s_dev;')\n", dbConn, dbName)
			fmt.Printf("2. Add 'DB_DEV_NAME=%s_dev' to your .env file\n", dbName)
			fmt.Println("3. Run the command again.")
		}
		os.Exit(1)
	}

	Info("Cleaning up SQL files...")
	files, _ := filepath.Glob("internal/database/migrations/*.sql")
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			Warning("Failed to read %s: %v", f, err)
			continue
		}
		newContent := string(content)
		newContent = strings.ReplaceAll(newContent, "`", "")

		lines := strings.Split(newContent, "\n")
		for i, line := range lines {
			if idx := strings.Index(line, "COLLATE"); idx != -1 {
				lines[i] = strings.TrimSpace(line[:idx])
				if strings.HasSuffix(lines[i], ";") {
				} else {
					lines[i] += ";"
				}
			}
		}
		newContent = strings.Join(lines, "\n")

		if err := os.WriteFile(f, []byte(newContent), 0644); err != nil {
			Warning("Failed to write %s: %v", f, err)
		}
	}

	Info("Running atlas migrate hash...")
	cmd = exec.Command("atlas", "migrate", "hash", "--env", "gorm")
	cmd.Env = atlasEnv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		ErrorLog("Atlas hash failed: %v", err)
		os.Exit(1)
	}

	Success("Migration created successfully")
}

func runLoader() {
	cfg, _ := env.Load()
	moduleName := "github.com/mmycin/goforge"
	dbConn := "sqlite"
	if cfg != nil {
		if cfg.Module != "" {
			moduleName = cfg.Module
		}
		if cfg.DBConnection != "" {
			dbConn = cfg.DBConnection
		}
	}

	loaderSource := fmt.Sprintf(`//go:build ignore
package main

import (
	"fmt"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	"%s/internal/services"
)

func main() {
	models := services.Model()
	loader := gormschema.New("%s")
	stmts, err := loader.Load(models...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load GORM schema: %%v\n", err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stdout, stmts)
}
`, moduleName, dbConn)

	tmpFile := ".goforge-tmp-loader.go"
	if err := os.WriteFile(tmpFile, []byte(loaderSource), 0644); err != nil {
		ErrorLog("Failed to create temporary loader file: %v", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFile)

	cmd := exec.Command("go", "run", tmpFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		ErrorLog("Failed to execute loader: %v", err)
		os.Exit(1)
	}
}

func migrateDB() {
	// The universal CLI can't easily auto-migrate local database via direct GORM
	// without executing code. Therefore we'll proxy it to "go run cmd/main.go migrate" if possible,
	// or we just run atlas apply.

	Info("Running Atlas migrate apply...")

	cfg, err := env.Load()
	dbConn := "sqlite"
	dbName := ""
	dbUser := ""
	dbPass := ""
	dbHost := ""
	dbPort := "3306"
	dbDevName := ""

	if err == nil && cfg != nil {
		if cfg.DBConnection != "" {
			dbConn = cfg.DBConnection
		}
		if cfg.DBName != "" {
			dbName = cfg.DBName
		}
		if cfg.DBUsername != "" {
			dbUser = cfg.DBUsername
		}
		if cfg.DBPassword != "" {
			dbPass = cfg.DBPassword
		}
		if cfg.DBHost != "" {
			dbHost = cfg.DBHost
		}
		if cfg.DBPort != "" {
			dbPort = cfg.DBPort
		}
		if cfg.DBDevName != "" {
			dbDevName = cfg.DBDevName
		}
	}

	atlasEnv := os.Environ()
	atlasEnv = append(atlasEnv, "DB_CONNECTION="+dbConn)
	atlasEnv = append(atlasEnv, "DB_NAME="+dbName)
	atlasEnv = append(atlasEnv, "DB_USERNAME="+dbUser)
	atlasEnv = append(atlasEnv, "DB_PASSWORD="+dbPass)
	atlasEnv = append(atlasEnv, "DB_HOST="+dbHost)
	atlasEnv = append(atlasEnv, "DB_PORT="+dbPort)
	atlasEnv = append(atlasEnv, "DB_DEV_NAME="+dbDevName)

	cmd := exec.Command("atlas", "migrate", "apply", "--env", "gorm")
	cmd.Env = atlasEnv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		ErrorLog("Atlas migrate apply failed: %v", err)
		os.Exit(1)
	}
	Success("Atlas migration completed successfully")
}

