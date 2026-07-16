package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/mmycin/GoForge/internal/infra"
)

// Generator orchestrates template rendering and file writing.
// It depends only on the FileSystem interface — no real os calls.
type Generator struct {
	fs infra.FileSystem
}

// NewGenerator returns a Generator that writes through the given FileSystem.
func NewGenerator(fs infra.FileSystem) *Generator {
	return &Generator{fs: fs}
}

// Generate renders the template identified by key with data and writes the
// result to destPath, creating parent directories as needed.
func (g *Generator) Generate(key string, data any, destPath string) error {
	content, err := RenderKey(key, data)
	if err != nil {
		return fmt.Errorf("render %q → %s: %w", key, destPath, err)
	}

	// filepath.Dir handles both / and \ separators correctly on all platforms.
	dir := filepath.Dir(destPath)
	if dir != "" && dir != "." {
		if err := g.fs.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}

	return g.fs.WriteFile(destPath, []byte(content), 0644)
}

// Exists reports whether path already exists on the file system.
func (g *Generator) Exists(path string) bool {
	_, err := g.fs.Stat(path)
	return err == nil
}

// ToCamelCase converts snake_case or kebab-case to CamelCase.
// Exported so templates can call it via the "title" func map.
func ToCamelCase(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		r := []rune(part)
		r[0] = unicode.ToUpper(r[0])
		parts[i] = string(r)
	}
	return strings.Join(parts, "")
}

// SafeIdentifier converts a command name like "emails:send-digest" into a
// safe Go package/variable name like "emails_send_digest".
func SafeIdentifier(name string) string {
	r := strings.NewReplacer(":", "_", "-", "_", ".", "_", "/", "_")
	return r.Replace(name)
}

// EnvPrefix converts snake_case "mail_server" to "MAIL_SERVER".
func EnvPrefix(name string) string {
	return strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
}

// ListServices reads internal/services/ via the file system and returns the
// names of directories that contain a model.go file.
func (g *Generator) ListServices() ([]string, error) {
	entries, err := g.fs.ReadDir("internal/services")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			_, statErr := g.fs.Stat("internal/services/" + e.Name() + "/model.go")
			if statErr == nil {
				names = append(names, e.Name())
			}
		}
	}
	return names, nil
}
