package cmd

import (
	"fmt"
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
		"handler.go": fmt.Sprintf(`package %s

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type %sHandler struct {}

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
`, name, camelName, camelName, camelName),
		"grpc.go": "package " + name + "\n",
		"routes.go": fmt.Sprintf(`package %s

import (
	"github.com/gin-gonic/gin"
	"%s/internal/server"
)

func init() {
	server.Register(&%sRoutes{})
}

type %sRoutes struct{}

func (r *%sRoutes) Register(engine *gin.Engine) {
	h := &%sHandler{}

	group := engine.Group("/api/%ss")
	// Middleware is applied globally in server/http.go

	group.GET("/", h.GetAll)
	group.GET("/:id", h.GetByID)
}
`, name, moduleName, camelName, camelName, camelName, camelName, name),
		"docs.go": fmt.Sprintf(`package %s

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"%s/internal/server"
)

func init() {
	server.Register(&%sDocs{})
}

type %sDocs struct{}

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
`, name, moduleName, camelName, camelName, camelName, name, name, name),
		"service.go": "package " + name + "\n",
		"model.go":   fmt.Sprintf("package %s\n\nimport \"time\"\n\ntype %s struct {\n\tID        uint      `gorm:\"primaryKey;autoIncrement\" json:\"id\"`\n\tCreatedAt time.Time `gorm:\"autoCreateTime\" json:\"created_at\"`\n\tUpdatedAt time.Time `gorm:\"autoUpdateTime\" json:\"updated_at\"`\n}\n", name, camelName),
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

	var services []string
	for _, e := range entries {
		if e.IsDir() {
			modelPath := filepath.Join(servicesDir, e.Name(), "model.go")
			if _, err := os.Stat(modelPath); err == nil {
				services = append(services, e.Name())
			}
		}
	}

	tmpl := `package services

import (
	"{{ .Module }}/internal/server"
{{- range .Services }}
	"{{ $.Module }}/internal/services/{{ . }}"
{{- end }}
)

// GetRouters returns all service routers to be registered
func GetRouters() []server.Router {
	return server.GetRegisteredRouters()
}

// Model returns all models to be registered with GORM
func Model() []any {
	return []any{
{{- range .Services }}
		&{{ . }}.{{ title . }}{},
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
		Services []string
	}{
		Module:   moduleName,
		Services: services,
	}

	return t.Execute(f, data)
}
