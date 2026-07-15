package cmd

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/mmycin/GoForge/internal/env"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(makeServiceCmd)
}

var makeServiceCmd = &cobra.Command{
	Use:   "gen:service [name]",
	Short: "Create a new service",
	Long:  `Generate a new service with handler, repository, model, routes, and proto files.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		Info("Creating service: %s", name)
		genService(name)
	},
}

func genService(name string) {
	targetDir := filepath.Join("internal/services", name)

	if _, err := os.Stat(targetDir); err == nil {
		ErrorLog("Service '%s' already exists", name)
		os.Exit(1)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		ErrorLog("Failed to create directory: %v", err)
		os.Exit(1)
	}

	camelName := toCamelCase(name)

	// Load the module name from local environment
	cfg, err := env.Load()
	moduleName := "github.com/mmycin/goforge"
	if err == nil && cfg.Module != "" {
		moduleName = cfg.Module
	} else {
		Warning("Could not read module from go.mod, using default: %s", moduleName)
	}

	files := map[string]string{
		"service.go": fmt.Sprintf(`package %s

import (
	"context"

	"%s/internal/database"
)

type %sService struct {
	db *database.Database
}

func New%sService(db *database.Database) *%sService {
	return &%sService{db: db}
}
`, name, moduleName, camelName, camelName, camelName, camelName),
		"handler.go": fmt.Sprintf(`package %s

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type %sHandler struct {
	service *%sService
}

func New%sHandler(service *%sService) *%sHandler {
	return &%sHandler{service: service}
}

func (h *%sHandler) GetAll(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Data retrieved",
		"data":    []string{},
	})
}

func (h *%sHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Detail retrieved",
		"data":    id,
	})
}
`, name, camelName, camelName, camelName, camelName, camelName, camelName, camelName),
		"grpc.go": "package " + name + "\n",
		"routes.go": fmt.Sprintf(`package %s

import (
	"github.com/gin-gonic/gin"
)

type %sRoutes struct {
	handler *%sHandler
}

func New%sRoutes(handler *%sHandler) *%sRoutes {
	return &%sRoutes{handler: handler}
}

func (r *%sRoutes) Register(engine *gin.Engine) {
	group := engine.Group("/api/%ss")
	// Middleware is applied globally in server/http.go

	group.GET("/", r.handler.GetAll)
	group.GET("/:id", r.handler.GetByID)
}
`, name, camelName, camelName, camelName, camelName, camelName, camelName, name),
		"docs.go": fmt.Sprintf(`package %s

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"%s/boot/server"
)

type %sDocs struct{}

func New%sDocs() *%sDocs {
	return &%sDocs{}
}

func (d *%sDocs) Register(engine *gin.Engine) {
	config := server.NewHumaConfig("%s API", "1.0.0", "/api/docs/%s")
	
	// Create API instance
	api := humagin.New(engine, config)
	
	// Register health check
	huma.Register(api, huma.Operation{
		OperationID: "get-health",
		Method:      http.MethodGet,
		Path:        "/api/%s/health",
		Summary:     "Health check",
		Description: "Check if the service is healthy",
		Tags:        []string{"Health"},
	}, func(ctx context.Context, input *struct{}) (*struct{ Body string }, error) {
		return &struct{ Body string }{Body: "OK"}, nil
	})
}
`, name, moduleName, camelName, camelName, camelName, camelName, camelName, name, name, name),
		"model.go": fmt.Sprintf("package %s\n\nimport \"time\"\n\ntype %s struct {\n\tID        uint      `gorm:\"primaryKey;autoIncrement\" json:\"id\"`\n\tCreatedAt time.Time `gorm:\"autoCreateTime\" json:\"created_at\"`\n\tUpdatedAt time.Time `gorm:\"autoUpdateTime\" json:\"updated_at\"`\n}\n\nfunc (t *%s) To%sModel() *%s {\n\treturn &%s{\n\t\tID:        t.ID,\n\t\tCreatedAt: t.CreatedAt,\n\t\tUpdatedAt: t.UpdatedAt,\n\t}\n}\n", name, camelName, camelName, camelName, camelName, camelName),
	}

	for fname, content := range files {
		if err := os.WriteFile(filepath.Join(targetDir, fname), []byte(content), 0644); err != nil {
			ErrorLog("Failed to write %s: %v", fname, err)
		}
	}

	// Create proto directory and file
	protoDir := filepath.Join("proto", name)
	if err := os.MkdirAll(protoDir, 0755); err != nil {
		ErrorLog("Failed to create proto directory: %v", err)
	} else {
		// Proto content stays similar but ensure package name is simple
		protoContent := fmt.Sprintf(`syntax = "proto3";

package %s;

option go_package = "%s/proto/%s/gen";

service %sService {
	rpc Create(CreateRequest) returns (CreateResponse);
	rpc Get(GetRequest) returns (GetResponse);
	rpc List(ListRequest) returns (ListResponse);
	rpc Update(UpdateRequest) returns (UpdateResponse);
	rpc Delete(DeleteRequest) returns (DeleteResponse);
}

message %s {
	string id = 1;
	string created_at = 2;
	string updated_at = 3;
}

message CreateRequest {}
message CreateResponse {}

message GetRequest { string id = 1; }
message GetResponse {}

message ListRequest { int32 page = 1; int32 limit = 2; }
message ListResponse {}

message UpdateRequest { string id = 1; }
message UpdateResponse {}

message DeleteRequest { string id = 1; }
message DeleteResponse {}
`, name, moduleName, name, camelName, camelName)

		if err := os.WriteFile(filepath.Join(protoDir, name+".proto"), []byte(protoContent), 0644); err != nil {
			ErrorLog("Failed to write proto file: %v", err)
		}
	}

	if err := registerModels(moduleName); err != nil {
		Warning("Could not automatically update kernel.go: %v", err)
	}

	Success("Service '%s' created successfully and auto-registered in kernel.go", name)
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			r := []rune(part)
			r[0] = unicode.ToUpper(r[0])
			parts[i] = string(r)
		}
	}
	return strings.Join(parts, "")
}

func registerModels(moduleName string) error {
	servicesDir := "internal/services"
	entries, err := os.ReadDir(servicesDir)
	if err != nil {
		return err
	}

	type ServiceInfo struct {
		Name   string
		Models []string
	}

	var services []ServiceInfo
	for _, e := range entries {
		if e.IsDir() {
			modelPath := filepath.Join(servicesDir, e.Name(), "model.go")
			if _, err := os.Stat(modelPath); err == nil {
				models, err := findModelsInFile(modelPath)
				if err != nil {
					Warning("Could not parse models in %s: %v", modelPath, err)
					// Fallback to title case of service name if parsing fails
					models = []string{toCamelCase(e.Name())}
				}
				if len(models) > 0 {
					services = append(services, ServiceInfo{
						Name:   e.Name(),
						Models: models,
					})
				}
			}
		}
	}

	tmpl := `package services

import (
	"{{ .Module }}/boot/server"
	"{{ .Module }}/internal/database"
{{- range .Services }}
	"{{ $.Module }}/internal/services/{{ .Name }}"
{{- end }}
)

// ServicesConfig contains all dependencies needed to initialize services
type ServicesConfig struct {
	DB *database.Database
}

// InitializeServices initializes all services and returns routers and gRPC registries
func InitializeServices(cfg ServicesConfig) ([]server.Router, []server.GRPCRegistry) {
	var routers []server.Router
	var grpcRegistries []server.GRPCRegistry
{{- range .Services }}
	// Initialize {{ .Name }} service
	{{ .Name }}Service := {{ .Name }}.New{{ title .Name }}Service(cfg.DB)
	{{ .Name }}Handler := {{ .Name }}.New{{ title .Name }}Handler({{ .Name }}Service)
	{{ .Name }}Routes := {{ .Name }}.New{{ title .Name }}Routes({{ .Name }}Handler)
	{{ .Name }}Docs := {{ .Name }}.New{{ title .Name }}Docs()
	routers = append(routers, {{ .Name }}Routes, {{ .Name }}Docs)
{{- end }}
	return routers, grpcRegistries
}

// Model returns all models to be registered with GORM
func Model() []any {
	return []any{
{{- range $service := .Services }}
	{{- range .Models }}
		&{{ $service.Name }}.{{ . }}{},
	{{- end }}
{{- end }}
	}
}
`
	funcMap := template.FuncMap{
		"title": func(s string) string {
			parts := strings.Split(s, "_")
			for i, part := range parts {
				if len(part) > 0 {
					r := []rune(part)
					r[0] = unicode.ToUpper(r[0])
					parts[i] = string(r)
				}
			}
			return strings.Join(parts, "")
		},
	}

	t, err := template.New("kernel").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create("internal/services/kernel.go")
	if err != nil {
		return err
	}
	defer f.Close()

	data := struct {
		Module   string
		Services []ServiceInfo
	}{
		Module:   moduleName,
		Services: services,
	}

	return t.Execute(f, data)
}

func findModelsInFile(path string) ([]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}

	var models []string
	for _, f := range node.Decls {
		genDecl, ok := f.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			// Check if it's a struct and exported
			if _, ok := typeSpec.Type.(*ast.StructType); ok && ast.IsExported(typeSpec.Name.Name) {
				models = append(models, typeSpec.Name.Name)
			}
		}
	}
	return models, nil
}
