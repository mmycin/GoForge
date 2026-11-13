package usecase

import (
	"fmt"
	"strings"

	"github.com/mmycin/GoForge/internal/infra"
)

// addEntryToModuleSlice inserts symbol into a slice field (e.g. "Providers" or
// "Invokes") inside a service's module.go file. It is idempotent.
func addEntryToModuleSlice(fsys infra.FileSystem, moduleFilePath, sliceField, symbol string) (bool, error) {
	raw, err := fsys.ReadFile(moduleFilePath)
	if err != nil {
		return false, err
	}
	src := string(raw)

	if strings.Contains(src, symbol) {
		return false, nil
	}

	emptyLiteral := sliceField + ": []any{}"
	if strings.Contains(src, emptyLiteral) {
		replacement := fmt.Sprintf("%s: []any{\n\t\t\t%s,\n\t\t}", sliceField, symbol)
		updated := strings.Replace(src, emptyLiteral, replacement, 1)
		return true, fsys.WriteFile(moduleFilePath, []byte(updated), 0644)
	}

	openLiteral := sliceField + ": []any{"
	idx := strings.Index(src, openLiteral)
	if idx == -1 {
		return false, fmt.Errorf("could not locate %s slice in %s", sliceField, moduleFilePath)
	}

	insertPos := idx + len(openLiteral)
	updated := src[:insertPos] + "\n\t\t\t" + symbol + "," + src[insertPos:]
	return true, fsys.WriteFile(moduleFilePath, []byte(updated), 0644)
}

// removeEntryFromModuleFile removes any line containing symbol from module.go.
func removeEntryFromModuleFile(fsys infra.FileSystem, moduleFilePath, symbol string) (bool, error) {
	raw, err := fsys.ReadFile(moduleFilePath)
	if err != nil {
		return false, err
	}
	lines := strings.Split(string(raw), "\n")
	out := make([]string, 0, len(lines))
	changed := false
	for _, line := range lines {
		if strings.Contains(line, symbol) {
			changed = true
			continue
		}
		out = append(out, line)
	}
	if !changed {
		return false, nil
	}
	return true, fsys.WriteFile(moduleFilePath, []byte(strings.Join(out, "\n")), 0644)
}

