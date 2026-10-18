package usecase

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mmycin/GoForge/internal/infra"
)

// RemGRPCUseCase removes all gRPC infrastructure from a goforge-template project,
// leaving a pure REST-only codebase.
type RemGRPCUseCase struct {
	fs   infra.FileSystem
	exec infra.Executor
}

// NewRemGRPCUseCase constructs a RemGRPCUseCase.
func NewRemGRPCUseCase(fs infra.FileSystem, exec infra.Executor) *RemGRPCUseCase {
	return &RemGRPCUseCase{fs: fs, exec: exec}
}

// Run removes every gRPC artifact from the project.
// It is idempotent — missing files are treated as warnings, not errors.
func (uc *RemGRPCUseCase) Run(progress Progress) error {
	// ── 1. boot/server/grpc.go ──────────────────────────────────────────────
	uc.removeFile("boot/server/grpc.go", progress)

	// ── 2. boot/client/grpc_client.go ───────────────────────────────────────
	uc.removeFile("boot/client/grpc_client.go", progress)

	// ── 3. boot/server/middleware/grpc.go ───────────────────────────────────
	uc.removeFile("boot/server/middleware/grpc.go", progress)

	// ── 4. core/config/grpc.go ──────────────────────────────────────────────
	uc.removeFile("core/config/grpc.go", progress)

	// ── 5. Patch core/config/config.go ──────────────────────────────────────
	send(progress, StepStarted{Label: "Patching core/config/config.go"})
	if err := uc.patchConfigGo(); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("config.go: %v", err)})
	} else {
		send(progress, StepDone{Label: "Patched core/config/config.go"})
	}

	// ── 6. Remove grpc.go stub from every service ───────────────────────────
	send(progress, StepStarted{Label: "Removing per-service grpc.go stubs"})
	entries, _ := uc.fs.ReadDir("internal/services")
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p := filepath.Join("internal/services", e.Name(), "grpc.go")
		if _, err := uc.fs.Stat(p); err == nil {
			if err := uc.fs.Remove(p); err != nil {
				send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("remove %s: %v", p, err)})
			}
		}
	}
	send(progress, StepDone{Label: "Removed per-service grpc.go stubs"})

	// ── 7. Remove internal/proto/ contents ──────────────────────────────────
	send(progress, StepStarted{Label: "Removing internal/proto contents"})
	protoEntries, _ := uc.fs.ReadDir("internal/proto")
	for _, e := range protoEntries {
		if e.Name() == ".gitkeep" {
			continue
		}
		p := filepath.Join("internal/proto", e.Name())
		if err := uc.fs.RemoveAll(p); err != nil {
			send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("remove %s: %v", p, err)})
		}
	}
	send(progress, StepDone{Label: "Removed internal/proto contents"})

	// ── 8. Remove internal/tests/grpc/ ──────────────────────────────────────
	send(progress, StepStarted{Label: "Removing internal/tests/grpc"})
	grpcTestEntries, _ := uc.fs.ReadDir("internal/tests/grpc")
	for _, e := range grpcTestEntries {
		if e.Name() == ".gitkeep" {
			continue
		}
		p := filepath.Join("internal/tests/grpc", e.Name())
		if err := uc.fs.Remove(p); err != nil {
			send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("remove %s: %v", p, err)})
		}
	}
	send(progress, StepDone{Label: "Removed internal/tests/grpc"})

	// ── 9. Patch internal/services/kernel.go ────────────────────────────────
	send(progress, StepStarted{Label: "Patching internal/services/kernel.go"})
	if err := uc.patchServicesKernel(); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("kernel.go: %v", err)})
	} else {
		send(progress, StepDone{Label: "Patched internal/services/kernel.go"})
	}

	// ── 10. Patch core/console/serve.go ─────────────────────────────────────
	send(progress, StepStarted{Label: "Patching core/console/serve.go"})
	if err := uc.patchServeGo(); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("serve.go: %v", err)})
	} else {
		send(progress, StepDone{Label: "Patched core/console/serve.go"})
	}

	// ── 11. Patch boot/kernel.go ─────────────────────────────────────────────
	send(progress, StepStarted{Label: "Patching boot/kernel.go"})
	if err := uc.patchBootKernel(); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("boot/kernel.go: %v", err)})
	} else {
		send(progress, StepDone{Label: "Patched boot/kernel.go"})
	}

	// ── 12. go mod tidy ─────────────────────────────────────────────────────
	send(progress, StepStarted{Label: "Running go mod tidy"})
	if err := uc.exec.Run("go", "mod", "tidy"); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("go mod tidy: %v", err)})
	} else {
		send(progress, StepDone{Label: "go mod tidy complete"})
	}

	send(progress, UseCaseDone{})
	return nil
}

// removeFile deletes a single file, emitting progress events.
func (uc *RemGRPCUseCase) removeFile(path string, progress Progress) {
	send(progress, StepStarted{Label: "Removing " + path})
	if _, err := uc.fs.Stat(path); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("%s not found, skipping", path)})
		return
	}
	if err := uc.fs.Remove(path); err != nil {
		send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("remove %s: %v", path, err)})
	} else {
		send(progress, StepDone{Label: "Removed " + path})
	}
}

// patchConfigGo removes the GRPC field, package-level var, and loadGRPCConfig call
// from core/config/config.go using string surgery.
func (uc *RemGRPCUseCase) patchConfigGo() error {
	const path = "core/config/config.go"
	raw, err := uc.fs.ReadFile(path)
	if err != nil {
		return err
	}
	src := string(raw)

	// Remove struct field line
	src = removeLineContaining(src, "GRPC    GRPCConfig")

	// Remove package-level var line
	src = removeLineContaining(src, "GRPC    GRPCConfig")
	// (same token — handle both by running twice; second run is a no-op if gone)
	src = removeLineContaining(src, "GRPC GRPCConfig")

	// Remove the loadGRPCConfig block (3 lines pattern):
	//   if cfg.GRPC, err = loadGRPCConfig(); err != nil {
	//       return nil, err
	//   }
	src = removeBlockStartingWith(src, "\tif cfg.GRPC, err = loadGRPCConfig()")

	// Remove package-level assignment: GRPC = cfg.GRPC
	src = removeLineContaining(src, "GRPC = cfg.GRPC")

	// Remove package-level var declaration: GRPC    GRPCConfig
	// (already handled above, but run once more to catch spacing variants)
	src = removeLineContaining(src, "GRPC     GRPCConfig")

	return uc.fs.WriteFile(path, []byte(src), 0644)
}

// patchServicesKernel removes GRPCRegistry from the return signature and
// the grpcRegistries variable from internal/services/kernel.go.
func (uc *RemGRPCUseCase) patchServicesKernel() error {
	const path = "internal/services/kernel.go"
	raw, err := uc.fs.ReadFile(path)
	if err != nil {
		return err
	}
	src := string(raw)

	// Change function signature return type
	src = strings.ReplaceAll(src, "([]server.Router, []server.GRPCRegistry)", "[]server.Router")

	// Remove grpcRegistries variable declaration
	src = removeLineContaining(src, "grpcRegistries := []server.GRPCRegistry{}")

	// Remove grpcRegistries return value from return statements
	// Pattern: "return routers, grpcRegistries" → "return routers"
	src = strings.ReplaceAll(src, "return routers, grpcRegistries", "return routers")

	// Remove any lingering grpc import if present
	src = removeLineContaining(src, `"google.golang.org/grpc"`)

	// Remove server.GRPCRegistry import hint from function comment if it exists
	// (cosmetic — won't break anything if not found)

	return uc.fs.WriteFile(path, []byte(src), 0644)
}

// patchServeGo removes the gRPC server construction block from core/console/serve.go.
func (uc *RemGRPCUseCase) patchServeGo() error {
	const path = "core/console/serve.go"
	raw, err := uc.fs.ReadFile(path)
	if err != nil {
		return err
	}
	src := string(raw)

	// Remove the grpcRegistries variable from the InitializeServices destructuring.
	// "routers, grpcRegistries := services.InitializeServices" → "routers := services.InitializeServices"
	// But InitializeServices may already return one value after the kernel patch,
	// so handle both forms gracefully.
	src = strings.ReplaceAll(src,
		"routers, grpcRegistries := services.InitializeServices",
		"routers := services.InitializeServices")

	// Remove the gRPC server construction block:
	//   if cfg.GRPC.Enable {
	//       servers = append(servers, server.NewGRPCServer(cfg, grpcRegistries))
	//   }
	src = removeBlockStartingWith(src, "\t\tif cfg.GRPC.Enable {")
	// Alternative form (no extra tab indent)
	src = removeBlockStartingWith(src, "\tif cfg.GRPC.Enable {")

	// Remove standalone grpcRegistries append line if it somehow survived
	src = removeLineContaining(src, "server.NewGRPCServer(")

	// Remove "boot/server" import line if GRPCServer was the only server import
	// (safe to leave — Go compiler will warn if unused, but we keep it for Router)

	// Remove grpc-related import lines
	src = removeLineContaining(src, `"google.golang.org/grpc"`)

	return uc.fs.WriteFile(path, []byte(src), 0644)
}

// patchBootKernel rewrites boot/kernel.go to remove all gRPC references:
//   - drops grpcRegistries param from NewApp
//   - removes the buildServers helper (which held the GRPC.Enable conditional)
//   - replaces the buildServers call with a direct HTTP-only server slice
func (uc *RemGRPCUseCase) patchBootKernel() error {
	const path = "boot/kernel.go"
	raw, err := uc.fs.ReadFile(path)
	if err != nil {
		return err
	}
	src := string(raw)

	// 1. Remove grpcRegistries parameter from NewApp signature.
	//    Handles both comma-leading and comma-trailing variants.
	src = strings.ReplaceAll(src, ", grpcRegistries []server.GRPCRegistry", "")
	src = strings.ReplaceAll(src, "grpcRegistries []server.GRPCRegistry, ", "")
	src = strings.ReplaceAll(src, "grpcRegistries []server.GRPCRegistry", "")

	// 2. Replace buildServers call with a direct HTTP-only slice.
	//    Pattern: servers := buildServers(cfg, httpRouters, grpcRegistries)
	src = strings.ReplaceAll(src,
		"servers := buildServers(cfg, httpRouters, grpcRegistries)",
		"servers := []server.Server{\n\t\tserver.NewHTTPServer(cfg, httpRouters),\n\t}",
	)
	// Fallback: buildServers with only two args (if already partially patched)
	src = strings.ReplaceAll(src,
		"servers := buildServers(cfg, httpRouters)",
		"servers := []server.Server{\n\t\tserver.NewHTTPServer(cfg, httpRouters),\n\t}",
	)

	// 3. Remove the entire buildServers function.
	src = removeFuncNamed(src, "buildServers")

	// 4. Remove any leftover blank lines caused by the removals (3+ consecutive newlines → 2).
	for strings.Contains(src, "\n\n\n") {
		src = strings.ReplaceAll(src, "\n\n\n", "\n\n")
	}

	return uc.fs.WriteFile(path, []byte(src), 0644)
}

// removeFuncNamed removes a top-level Go function by name using brace counting.
func removeFuncNamed(src, name string) string {
	// Match "func <name>(" at the start of a line.
	marker := "func " + name + "("
	idx := strings.Index(src, marker)
	if idx == -1 {
		return src
	}
	// Walk backwards to include any preceding comment block.
	start := idx
	// Find preceding newline to check for comment lines.
	pre := src[:idx]
	lines := strings.Split(pre, "\n")
	// Walk backward through comment lines
	trimIdx := len(lines) - 1
	for trimIdx > 0 && (strings.HasPrefix(strings.TrimSpace(lines[trimIdx-1]), "//") || strings.TrimSpace(lines[trimIdx-1]) == "") {
		trimIdx--
	}
	if trimIdx < len(lines)-1 {
		start = len(strings.Join(lines[:trimIdx+1], "\n")) + 1 // +1 for the newline
	}

	// Scan forward to find the closing brace.
	depth := 0
	end := idx
	foundOpen := false
	for i := idx; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
			foundOpen = true
		case '}':
			depth--
			if foundOpen && depth == 0 {
				end = i + 1
				goto done
			}
		}
	}
done:
	// Consume the trailing newline after the closing brace.
	if end < len(src) && src[end] == '\n' {
		end++
	}

	return src[:start] + src[end:]
}

// removeLineContaining removes the first line that contains the given substring.
func removeLineContaining(src, needle string) string {
	lines := strings.Split(src, "\n")
	out := make([]string, 0, len(lines))
	removed := false
	for _, l := range lines {
		if !removed && strings.Contains(l, needle) {
			removed = true
			continue
		}
		out = append(out, l)
	}
	return strings.Join(out, "\n")
}

// removeBlockStartingWith removes a brace-delimited block whose first line
// starts with the given prefix (after trimming leading whitespace is NOT done —
// the prefix must match the literal indentation in the file).
func removeBlockStartingWith(src, prefix string) string {
	lines := strings.Split(src, "\n")
	out := make([]string, 0, len(lines))
	i := 0
	for i < len(lines) {
		if strings.HasPrefix(lines[i], prefix) || strings.TrimSpace(lines[i]) == strings.TrimSpace(prefix) {
			// Skip lines until the closing brace at the same indent level.
			depth := 0
			for i < len(lines) {
				for _, ch := range lines[i] {
					if ch == '{' {
						depth++
					} else if ch == '}' {
						depth--
					}
				}
				i++
				if depth == 0 && i > 0 {
					break
				}
			}
			continue
		}
		out = append(out, lines[i])
		i++
	}
	return strings.Join(out, "\n")
}
