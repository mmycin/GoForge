package scaffold

import (
	"bytes"
	"fmt"
	"text/template"
)

// Render executes tmpl with data and returns the rendered string.
// data must be one of the typed structs defined in data.go.
func Render(tmpl *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}
	return buf.String(), nil
}

// RenderKey is a convenience wrapper that loads and immediately renders
// the template identified by key with data.
func RenderKey(key string, data any) (string, error) {
	tmpl, err := LoadTemplate(key)
	if err != nil {
		return "", err
	}
	return Render(tmpl, data)
}
