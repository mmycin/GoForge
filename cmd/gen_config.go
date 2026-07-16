package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mmycin/GoForge/internal/env"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(genConfigCmd)
}

var genConfigCmd = &cobra.Command{
	Use:   "gen:config [name]",
	Short: "Generate a new config section",
	Long: `Scaffold a new config file in core/config/ and wire it into AllConfig.

Example:
  goforge gen:config cache
  → creates core/config/cache.go with CacheConfig struct + loadCacheConfig()
  → adds Cache CacheConfig field to AllConfig in core/config/config.go
  → adds the load call to Load() in core/config/config.go`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.ToLower(strings.TrimSpace(args[0]))
		if name == "" {
			ErrorLog("Config name cannot be empty")
			os.Exit(1)
		}
		Info("Generating config: %s", name)
		genConfig(name)
	},
}

func genConfig(name string) {
	configDir := filepath.Join("core", "config")

	// Guard: config dir must exist
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		ErrorLog("core/config/ does not exist — are you in the project root?")
		os.Exit(1)
	}

	// Resolve module name
	moduleName := "github.com/mmycin/goforge"
	if cfg, err := env.Load(); err == nil && cfg.Module != "" {
		moduleName = cfg.Module
	} else {
		Warning("Could not read module from go.mod, using default: %s", moduleName)
	}

	camel := toCamelCase(name)
	targetFile := filepath.Join(configDir, name+".go")

	// Guard: file must not already exist
	if _, err := os.Stat(targetFile); err == nil {
		ErrorLog("core/config/%s.go already exists", name)
		os.Exit(1)
	}

	writeConfigFile(targetFile, name, camel, moduleName)
	wireIntoAllConfig(name, camel)
}

// writeConfigFile creates core/config/<name>.go with the Config struct and loader.
func writeConfigFile(path, name, camel, moduleName string) {
	// Build example env key prefix from the name in upper-snake form.
	// e.g. "cache" → "CACHE", "mail_server" → "MAIL_SERVER"
	envPrefix := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))

	content := fmt.Sprintf(`package config

import "github.com/spf13/viper"

// %sConfig holds %s configuration.
// Add your fields here and map them to .env keys below.
type %sConfig struct {
	// Example fields — replace with your own:
	Enabled bool
	Host    string
	Port    int
}

func load%sConfig() (%sConfig, error) {
	return %sConfig{
		Enabled: viper.GetBool("%s_ENABLED"),
		Host:    viper.GetString("%s_HOST"),
		Port:    viper.GetInt("%s_PORT"),
	}, nil
}
`,
		camel, name,
		camel,
		camel, camel,
		camel,
		envPrefix,
		envPrefix,
		envPrefix,
	)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		ErrorLog("Failed to write %s: %v", path, err)
		os.Exit(1)
	}
	Success("Generated core/config/%s.go", name)
}

// wireIntoAllConfig patches core/config/config.go to add:
//  1. The field in AllConfig struct
//  2. The package-level var
//  3. The load call in Load()
//  4. The package-level assignment after Load()
func wireIntoAllConfig(name, camel string) {
	configPath := filepath.Join("core", "config", "config.go")
	raw, err := os.ReadFile(configPath)
	if err != nil {
		ErrorLog("Could not read core/config/config.go: %v", err)
		os.Exit(1)
	}
	src := string(raw)

	// Pad the field name to align with existing fields (longest is typically 7 chars)
	fieldPadded := fmt.Sprintf("%-7s %sConfig", camel, camel)
	fieldLine := "\t" + fieldPadded
	varPadded := fmt.Sprintf("%-7s %sConfig", camel, camel)
	varLine := "\t" + varPadded
	loadBlock := fmt.Sprintf("\tif cfg.%s, err = load%sConfig(); err != nil {\n\t\treturn nil, err\n\t}", camel, camel)
	assignLine := fmt.Sprintf("\t%s = cfg.%s", camel, camel)

	// ── 1. AllConfig struct field ─────────────────────────────────────────────
	if !strings.Contains(src, fieldLine) {
		// Insert before the closing brace of the AllConfig struct
		src = insertBeforeMarker(src, "\n}", "type AllConfig struct {", "\n"+fieldLine)
		if !strings.Contains(src, fieldLine) {
			Warning("Could not auto-wire AllConfig field — add manually:\n  %s", fieldLine)
		}
	}

	// ── 2. Package-level var ──────────────────────────────────────────────────
	if !strings.Contains(src, varLine) {
		// Find the var ( ... ) block and append before its closing paren
		src = insertBeforeMarker(src, "\n)", "var (", "\n"+varLine)
		if !strings.Contains(src, varLine) {
			Warning("Could not auto-wire package var — add manually:\n  var %s", varLine)
		}
	}

	// ── 3. Load call inside Load() ────────────────────────────────────────────
	if !strings.Contains(src, "load"+camel+"Config()") {
		// Insert the load block right before "// Set package-level variables" comment
		// or before the first package-level assignment block if comment is absent
		anchor := "\t// Set package-level variables"
		if !strings.Contains(src, anchor) {
			anchor = "\t" + "App = cfg.App"
		}
		if strings.Contains(src, anchor) {
			src = strings.Replace(src, anchor, loadBlock+"\n\n\t"+strings.TrimSpace(anchor), 1)
		} else {
			Warning("Could not auto-wire load call — add manually inside Load():\n  %s", loadBlock)
		}
	}

	// ── 4. Package-level assignment ───────────────────────────────────────────
	if !strings.Contains(src, assignLine) {
		// Insert after the last existing "X = cfg.X" line, before return
		anchor := "\treturn &cfg, nil"
		src = strings.Replace(src, anchor, assignLine+"\n\t"+strings.TrimSpace(anchor), 1)
		if !strings.Contains(src, assignLine) {
			Warning("Could not auto-wire assignment — add manually:\n  %s", assignLine)
		}
	}

	if err := os.WriteFile(configPath, []byte(src), 0644); err != nil {
		ErrorLog("Failed to update config.go: %v", err)
		os.Exit(1)
	}
	Success("Wired %sConfig into core/config/config.go", camel)
	Info("Add your .env keys:")
	envPrefix := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
	Info("  %s_ENABLED=false", envPrefix)
	Info("  %s_HOST=localhost", envPrefix)
	Info("  %s_PORT=0", envPrefix)
}

// insertBeforeMarker finds the first occurrence of scopeMarker, then finds the
// next occurrence of insertBefore after that, and inserts content before it.
func insertBeforeMarker(src, insertBefore, scopeMarker, content string) string {
	scopeIdx := strings.Index(src, scopeMarker)
	if scopeIdx == -1 {
		return src
	}
	// Search for insertBefore only within the block starting at scopeIdx
	after := src[scopeIdx:]
	relIdx := strings.Index(after, insertBefore)
	if relIdx == -1 {
		return src
	}
	absIdx := scopeIdx + relIdx
	return src[:absIdx] + content + src[absIdx:]
}
