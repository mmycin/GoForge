package usecase

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/mmycin/GoForge/internal/infra"
	"github.com/mmycin/GoForge/internal/scaffold"
)

// GenProtoUseCase compiles .proto files and scaffolds gRPC stubs.
type GenProtoUseCase struct {
	gen  *scaffold.Generator
	fs   infra.FileSystem
	exec infra.Executor
}

// NewGenProtoUseCase constructs a GenProtoUseCase.
func NewGenProtoUseCase(gen *scaffold.Generator, fs infra.FileSystem, exec infra.Executor) *GenProtoUseCase {
	return &GenProtoUseCase{gen: gen, fs: fs, exec: exec}
}

// Run compiles proto files for serviceName (empty = all services).
func (uc *GenProtoUseCase) Run(serviceName, modulePath string, progress Progress) error {
	protoDir := "internal/proto"

	var protoFiles []string
	if serviceName != "" {
		p := filepath.Join(protoDir, serviceName, serviceName+".proto")
		if !uc.gen.Exists(p) {
			return fmt.Errorf("proto file not found: %s", p)
		}
		protoFiles = append(protoFiles, p)
	} else {
		entries, err := uc.fs.ReadDir(protoDir)
		if err != nil {
			return fmt.Errorf("no proto directory found — run goforge gen:service first")
		}
		for _, e := range entries {
			if e.IsDir() {
				p := filepath.Join(protoDir, e.Name(), e.Name()+".proto")
				if uc.gen.Exists(p) {
					protoFiles = append(protoFiles, p)
				}
			}
		}
	}

	if len(protoFiles) == 0 {
		return fmt.Errorf("no proto files found")
	}

	for _, p := range protoFiles {
		svcName := filepath.Base(filepath.Dir(p))
		send(progress, StepStarted{Label: fmt.Sprintf("Compiling %s.proto", svcName)})

		var out bytes.Buffer
		args := []string{
			"--go_out=.", "--go_opt=module=" + modulePath,
			"--go-grpc_out=.", "--go-grpc_opt=module=" + modulePath,
			p,
		}
		if err := uc.exec.RunStreamed(context.Background(), &out, "protoc", args...); err != nil {
			send(progress, LogLine{Level: "error", Text: out.String()})
			send(progress, StepFailed{Label: fmt.Sprintf("Compile %s.proto", svcName), Err: err})
			return fmt.Errorf("protoc failed for %s: %w", p, err)
		}
		if out.Len() > 0 {
			for _, line := range strings.Split(out.String(), "\n") {
				if strings.TrimSpace(line) != "" {
					send(progress, LogLine{Level: "info", Text: line})
				}
			}
		}
		send(progress, StepDone{Label: fmt.Sprintf("Compiled %s.proto", svcName)})

		// Scaffold grpc.go
		send(progress, StepStarted{Label: fmt.Sprintf("Scaffolding grpc.go for %s", svcName)})
		if err := uc.scaffoldGrpc(svcName, modulePath); err != nil {
			send(progress, LogLine{Level: "warn", Text: fmt.Sprintf("grpc scaffold: %v", err)})
		} else {
			send(progress, StepDone{Label: fmt.Sprintf("Scaffolded grpc.go for %s", svcName)})
		}
	}

	send(progress, UseCaseDone{})
	return nil
}

func (uc *GenProtoUseCase) scaffoldGrpc(svcName, modulePath string) error {
	grpcFile := filepath.Join("internal/services", svcName, "grpc.go")
	methods, _ := uc.parseProtoMethods(svcName)

	data := scaffold.GrpcScaffoldData{
		PackageName: svcName,
		TypeName:    scaffold.ToCamelCase(svcName),
		ModulePath:  modulePath,
		Methods:     methods,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	return uc.gen.Generate("grpc/grpc.go.tmpl", data, grpcFile)
}

func (uc *GenProtoUseCase) parseProtoMethods(svcName string) ([]string, error) {
	protoFile := filepath.Join("internal/proto", svcName, svcName+".proto")
	content, err := uc.fs.ReadFile(protoFile)
	if err != nil {
		return nil, err
	}

	var methods []string
	inService := false
	for _, line := range strings.Split(string(content), "\n") {
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
				parts := strings.Fields(line)
				if len(parts) < 4 {
					continue
				}
				methodName := strings.Split(parts[1], "(")[0]
				reqPart := strings.Join(parts[1:], " ")
				rs, re := strings.Index(reqPart, "("), strings.Index(reqPart, ")")
				if rs == -1 || re == -1 {
					continue
				}
				reqType := reqPart[rs+1 : re]
				respPart := reqPart[re+1:]
				ss, se := strings.Index(respPart, "("), strings.Index(respPart, ")")
				if ss == -1 || se == -1 {
					continue
				}
				resType := respPart[ss+1 : se]
				methods = append(methods, fmt.Sprintf(
					"%s(ctx context.Context, in *pb.%s) (*pb.%s, error)",
					methodName, reqType, resType,
				))
			}
		}
	}
	return methods, nil
}
