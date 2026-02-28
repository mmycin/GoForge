package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/mmycin/GoForge/internal/env"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(genProtoCmd)
	rootCmd.AddCommand(remProtoCmd)
}

var genProtoCmd = &cobra.Command{
	Use:   "gen:proto [service]",
	Short: "Compile proto files and generate gRPC code",
	Long:  `Compiles .proto files using protoc and generates Go server scaffolding in internal/services.`,
	Run: func(cmd *cobra.Command, args []string) {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		if err := generateProto(name); err != nil {
			ErrorLog("%v", err)
			os.Exit(1)
		}
	},
}

var remProtoCmd = &cobra.Command{
	Use:   "rem:proto",
	Short: "Remove generated gRPC code",
	Long:  `Permanently remove generated .pb.go and _grpc.pb.go files.`,
	Run: func(cmd *cobra.Command, args []string) {
		Info("Removing generated proto files...")
		if err := removeProto(); err != nil {
			ErrorLog("%v", err)
			os.Exit(1)
		}
	},
}

func removeProto() error {
	servicesDir := "internal/services"
	entries, err := os.ReadDir(servicesDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			// Remove any .pb.go files in the service directory
			files, _ := filepath.Glob(filepath.Join(servicesDir, e.Name(), "*.pb.go"))
			for _, f := range files {
				Info("  Removing %s", f)
				os.Remove(f)
			}
			// Also look in gen/ if it exists
			genDir := filepath.Join("proto", e.Name(), "gen")
			if _, err := os.Stat(genDir); err == nil {
				Info("  Removing %s", genDir)
				os.RemoveAll(genDir)
			}
		}
	}

	Success("Generated proto files removed.")
	return nil
}

func generateProto(serviceName string) error {
	protoDir := "proto"
	servicesDir := "internal/services"

	// Find proto files
	var protoFiles []string
	if serviceName != "" {
		// Specific service
		p := filepath.Join(protoDir, serviceName, serviceName+".proto")
		if _, err := os.Stat(p); err == nil {
			protoFiles = append(protoFiles, p)
		} else {
			return fmt.Errorf("proto file not found: %s", p)
		}
	} else {
		// Scan all
		entries, err := os.ReadDir(protoDir)
		if err != nil {
			if os.IsNotExist(err) {
				Info("No proto directory found.")
				return nil
			}
			return err
		}
		for _, e := range entries {
			if e.IsDir() {
				p := filepath.Join(protoDir, e.Name(), e.Name()+".proto")
				if _, err := os.Stat(p); err == nil {
					protoFiles = append(protoFiles, p)
				}
			}
		}
	}

	if len(protoFiles) == 0 {
		Info("No proto files found.")
		return nil
	}

	cfg, _ := env.Load()
	moduleName := "github.com/mmycin/goforge"
	if cfg != nil && cfg.Module != "" {
		moduleName = cfg.Module
	} else {
		Warning("Could not determine module name: Using default: %s. Generated paths might be incorrect.", moduleName)
	}

	for _, p := range protoFiles {
		Info("Compiling %s...", p)

		args := []string{
			"--go_out=.",
			"--go-grpc_out=.",
			p,
		}

		if moduleName != "" {
			args = append(args, "--go_opt=module="+moduleName)
			args = append(args, "--go-grpc_opt=module="+moduleName)
		}

		cmd := exec.Command("protoc", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("protoc failed for %s: %w", p, err)
		}

		// Generate Scaffold
		// Extract service name from path: proto/<service>/<service>.proto
		parts := strings.Split(filepath.ToSlash(p), "/")
		if len(parts) >= 2 {
			svcName := parts[len(parts)-2]
			if err := generateGrpcScaffold(servicesDir, svcName, moduleName); err != nil {
				return err
			}
			if err := updateModelGo(svcName, moduleName); err != nil {
				Warning("  Could not update model.go for %s: %v", svcName, err)
			}
		}
	}

	Success("Proto compilation and scaffolding complete.")
	return nil
}

func updateModelGo(svcName, moduleName string) error {
	modelPath := filepath.Join("internal/services", svcName, "model.go")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return nil
	}

	content, err := os.ReadFile(modelPath)
	if err != nil {
		return err
	}

	src := string(content)
	if strings.Contains(src, ") ToProto()") || strings.Contains(src, ") ToModel()") {
		return nil
	}

	lines := strings.Split(src, "\n")
	var structName string
	var fields []string
	inStruct := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "type ") && strings.HasSuffix(trimmed, " struct {") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				structName = parts[1]
				inStruct = true
			}
			continue
		}
		if inStruct {
			if trimmed == "}" {
				inStruct = false
				break
			}
			if trimmed != "" && !strings.HasPrefix(trimmed, "//") {
				f := strings.Fields(trimmed)
				if len(f) >= 2 {
					// Check if it's a exported field
					if len(f[0]) > 0 && unicode.IsUpper(rune(f[0][0])) {
						fields = append(fields, f[0])
					}
				}
			}
		}
	}

	if structName == "" {
		return nil
	}

	var toProtoFields []string
	var toModelFields []string
	hasID := false

	for _, f := range fields {
		if f == "ID" {
			hasID = true
			toProtoFields = append(toProtoFields, "Id:          strconv.FormatUint(uint64(t.ID), 10),")
			toModelFields = append(toModelFields, "ID:          t.ID,")
		} else if f == "CreatedAt" || f == "UpdatedAt" {
			continue
		} else {
			toProtoFields = append(toProtoFields, fmt.Sprintf("%-12s t.%s,", f+":", f))
			toModelFields = append(toModelFields, fmt.Sprintf("%-12s t.%s,", f+":", f))
		}
	}

	methods := fmt.Sprintf(`
func (t *%s) ToProto() *pb.%s {
	return &pb.%s{
		%s
	}
}

func (t *%s) ToModel() *%s {
	return &%s{
		%s
	}
}
`, structName, structName, structName, strings.Join(toProtoFields, "\n\t\t"), structName, structName, structName, strings.Join(toModelFields, "\n\t\t"))

	// Add imports
	if !strings.Contains(src, "import (") && !strings.Contains(src, "import \"") {
		// No imports yet
		src = strings.Replace(src, "package "+svcName, "package "+svcName+"\n\nimport (\n)", 1)
	}

	if strings.Contains(src, "import (") {
		if hasID && !strings.Contains(src, "\"strconv\"") {
			src = strings.Replace(src, "import (", "import (\n\t\"strconv\"", 1)
		}
		if !strings.Contains(src, "\""+moduleName+"/proto/"+svcName+"/gen\"") {
			src = strings.Replace(src, "import (", fmt.Sprintf("import (\n\tpb \"%s/proto/%s/gen\"", moduleName, svcName), 1)
		}
	} else if strings.Contains(src, "import \"") {
		// Single import line, convert to block
		src = strings.Replace(src, "import \"", "import (\n\t\"", 1)
		src = strings.Replace(src, "\"\n", "\"\n)\n", 1)
		// Now it should have import ( and we can recurse or just handle it here
		if hasID && !strings.Contains(src, "\"strconv\"") {
			src = strings.Replace(src, "import (", "import (\n\t\"strconv\"", 1)
		}
		if !strings.Contains(src, "\""+moduleName+"/proto/"+svcName+"/gen\"") {
			src = strings.Replace(src, "import (", fmt.Sprintf("import (\n\tpb \"%s/proto/%s/gen\"", moduleName, svcName), 1)
		}
	}

	newContent := strings.TrimSpace(src) + "\n" + methods
	return os.WriteFile(modelPath, []byte(newContent), 0644)
}

func generateGrpcScaffold(servicesDir, serviceName, moduleName string) error {
	svcDir := filepath.Join(servicesDir, serviceName)
	if err := os.MkdirAll(svcDir, 0755); err != nil {
		return err
	}

	grpcFile := filepath.Join(svcDir, "grpc.go")
	Info("  Generating scaffold: %s", grpcFile)

	camelName := toCamelCase(serviceName)
	methods, err := parseProtoMethods(serviceName)
	if err != nil {
		Warning("  Could not parse proto file for stub generation: %v", err)
	}

	data := struct {
		Package   string
		Title     string
		Methods   []string
		Module    string
		Timestamp string
	}{
		Package:   serviceName,
		Title:     camelName,
		Methods:   methods,
		Module:    moduleName,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	tmpl := `// Generated at {{.Timestamp}}
package {{.Package}}

import (
	"context"

	"{{.Module}}/internal/server"
	pb "{{.Module}}/proto/{{.Package}}/gen"
	"google.golang.org/grpc"
)

type {{.Title}}GRPC struct {
	pb.Unimplemented{{.Title}}ServiceServer
}

func init() {
	server.RegisterGRPC(func(s *grpc.Server) {
		pb.Register{{.Title}}ServiceServer(s, &{{.Title}}GRPC{})
	})
}

{{range .Methods}}
func (t *{{$.Title}}GRPC) {{.}} {
	return nil, nil
}
{{end}}
`

	t, err := template.New("grpc").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(grpcFile)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, data)
}

func parseProtoMethods(serviceName string) ([]string, error) {
	protoFile := filepath.Join("proto", serviceName, serviceName+".proto")
	content, err := os.ReadFile(protoFile)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var methods []string
	inService := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "service ") {
			inService = true
			continue
		}
		if inService {
			if strings.HasPrefix(line, "}") {
				break
			}
			if strings.HasPrefix(line, "rpc ") {
				// rpc Create(CreateRequest) returns (CreateResponse);
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					methodName := parts[1]
					// Remove parentheses if name is like "Create("
					methodName = strings.Split(methodName, "(")[0]

					// Extract request type: (CreateRequest)
					reqPart := strings.Join(parts[1:], " ")
					reqStart := strings.Index(reqPart, "(")
					reqEnd := strings.Index(reqPart, ")")
					if reqStart != -1 && reqEnd != -1 {
						reqType := reqPart[reqStart+1 : reqEnd]

						// Extract response type: (CreateResponse)
						respPart := reqPart[reqEnd+1:]
						resStart := strings.Index(respPart, "(")
						resEnd := strings.Index(respPart, ")")
						if resStart != -1 && resEnd != -1 {
							resType := respPart[resStart+1 : resEnd]
							// Construct method signature
							methods = append(methods, fmt.Sprintf("%s(ctx context.Context, in *pb.%s) (*pb.%s, error)", methodName, reqType, resType))
						}
					}
				}
			}
		}
	}

	return methods, nil
}
