// Package scaffold provides typed data contracts for every code-generation
// template. All templates receive a concrete struct — never a map[string]any.
package scaffold

// ServiceData is passed to every template under templates/service/.
type ServiceData struct {
	PackageName string // snake_case, e.g. "product_catalog"
	TypeName    string // CamelCase, e.g. "ProductCatalog"
	ModulePath  string // full Go module path, e.g. "github.com/you/myapp"
}

// ProtoData is passed to templates/proto/service.proto.tmpl.
type ProtoData struct {
	PackageName string
	TypeName    string
	ModulePath  string
}

// GrpcScaffoldData is passed to templates/grpc/grpc.go.tmpl after proto compilation.
type GrpcScaffoldData struct {
	PackageName string
	TypeName    string
	ModulePath  string
	Methods     []string // pre-formatted Go method signatures
	Timestamp   string
}

// EventData is passed to both templates under templates/event/.
type EventData struct {
	PackageName string
	TypeName    string
	ModulePath  string
}

// CommandData is passed to templates/command/cmd.go.tmpl.
type CommandData struct {
	PackageName string // safe Go identifier, e.g. "emails_send_digest"
	CommandUse  string // cobra Use string, e.g. "emails:send-digest"
	TypeName    string // CamelCase struct name, e.g. "EmailsSendDigestCmd"
}

// ConfigData is passed to templates/config/config.go.tmpl.
type ConfigData struct {
	PackageName string // snake_case, e.g. "mailer"
	TypeName    string // CamelCase, e.g. "Mailer"
	EnvPrefix   string // UPPER_SNAKE, e.g. "MAILER"
	ModulePath  string
}

// KernelData is passed to the in-memory kernel.go template in make_service.
type KernelData struct {
	ModulePath string
	Services   []KernelService
}

// KernelService represents a single service entry in kernel.go.
type KernelService struct {
	Name   string   // snake_case package name
	Models []string // exported struct names found in model.go
}
