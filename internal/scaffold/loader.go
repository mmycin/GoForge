package scaffold

import (
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates
var templateFS embed.FS

// LoadTemplate parses and returns the template at the given key relative to
// the embedded templates/ directory.  key examples:
//
//	"service/service.go.tmpl"
//	"event/events.go.tmpl"
func LoadTemplate(key string) (*template.Template, error) {
	path := "templates/" + key
	content, err := templateFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("template %q not found in embedded FS: %w", key, err)
	}

	funcMap := template.FuncMap{
		// title converts snake_case to CamelCase — available inside every template.
		"title": ToCamelCase,
	}

	t, err := template.New(key).Funcs(funcMap).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %q: %w", key, err)
	}
	return t, nil
}
