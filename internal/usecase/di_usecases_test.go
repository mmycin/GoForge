package usecase

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mmycin/GoForge/internal/infra"
	"github.com/mmycin/GoForge/internal/scaffold"
)

func setupTempProject(t *testing.T) (string, infra.FileSystem, *scaffold.Generator) {
	t.Helper()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir %s: %v", tmpDir, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	fs := infra.NewOsFileSystem()
	if err := fs.MkdirAll("internal/services", 0755); err != nil {
		t.Fatalf("mkdir internal/services: %v", err)
	}
	return tmpDir, fs, scaffold.NewGenerator(fs)
}

func TestGenServiceUseCase_GeneratesModuleAndKernel(t *testing.T) {
	_, fs, gen := setupTempProject(t)
	uc := NewGenServiceUseCase(gen, fs)

	const modPath = "github.com/mmycin/goforge"
	if err := uc.Run("user", modPath, nil); err != nil {
		t.Fatalf("GenServiceUseCase(user): %v", err)
	}
	if err := uc.Run("product", modPath, nil); err != nil {
		t.Fatalf("GenServiceUseCase(product): %v", err)
	}

	userModuleRaw, err := fs.ReadFile("internal/services/user/module.go")
	if err != nil {
		t.Fatalf("read user/module.go: %v", err)
	}
	userModule := string(userModuleRaw)
	for _, want := range []string{
		"func Module() di.Module",
		`Name: "user"`,
		"NewUserService,",
		"NewUserHandler,",
		"NewUserRoutes,",
		"NewUserDocs,",
		"&User{},",
	} {
		if !strings.Contains(userModule, want) {
			t.Fatalf("user/module.go missing %q:\n%s", want, userModule)
		}
	}

	kernelRaw, err := fs.ReadFile("internal/services/kernel.go")
	if err != nil {
		t.Fatalf("read kernel.go: %v", err)
	}
	kernel := string(kernelRaw)
	for _, want := range []string{
		"func Modules() []di.Module",
		"product.Module(),",
		"user.Module(),",
		"func Model() []any",
		"return di.CollectModels(Modules()...)",
	} {
		if !strings.Contains(kernel, want) {
			t.Fatalf("kernel.go missing %q:\n%s", want, kernel)
		}
	}
	if strings.Contains(kernel, "NewUserService") || strings.Contains(kernel, "NewProductService") {
		t.Fatalf("kernel.go should not contain manual constructor calls:\n%s", kernel)
	}

	services, err := gen.ListServices()
	if err != nil {
		t.Fatalf("ListServices: %v", err)
	}
	if len(services) != 2 {
		t.Fatalf("expected 2 services from ListServices, got %v", services)
	}
}

func TestGenEventUseCase_PatchesModuleInvokesIdempotently(t *testing.T) {
	_, fs, gen := setupTempProject(t)
	svcUC := NewGenServiceUseCase(gen, fs)
	if err := svcUC.Run("user", "github.com/mmycin/goforge", nil); err != nil {
		t.Fatalf("GenServiceUseCase(user): %v", err)
	}

	evtUC := NewGenEventUseCase(gen, fs)
	if err := evtUC.Run("user", "github.com/mmycin/goforge", nil); err != nil {
		t.Fatalf("GenEventUseCase(user): %v", err)
	}

	listenersRaw, err := fs.ReadFile("internal/services/user/listeners.go")
	if err != nil {
		t.Fatalf("read listeners.go: %v", err)
	}
	if !strings.Contains(string(listenersRaw), "func RegisterListeners(") {
		t.Fatalf("listeners.go missing RegisterListeners:\n%s", string(listenersRaw))
	}

	modRaw, err := fs.ReadFile("internal/services/user/module.go")
	if err != nil {
		t.Fatalf("read module.go: %v", err)
	}
	if strings.Count(string(modRaw), "RegisterListeners") != 1 {
		t.Fatalf("expected RegisterListeners once in Invokes, got:\n%s", string(modRaw))
	}

	changed, err := addEntryToModuleSlice(fs, "internal/services/user/module.go", "Invokes", "RegisterListeners")
	if err != nil {
		t.Fatalf("addEntryToModuleSlice second call: %v", err)
	}
	if changed {
		t.Fatalf("expected addEntryToModuleSlice to be idempotent on duplicate symbol")
	}
}

func TestModulePatchHelpers_ProvidersAndRemoval(t *testing.T) {
	_, fs, gen := setupTempProject(t)
	svcUC := NewGenServiceUseCase(gen, fs)
	if err := svcUC.Run("order", "github.com/mmycin/goforge", nil); err != nil {
		t.Fatalf("GenServiceUseCase(order): %v", err)
	}

	modPath := filepath.Join("internal/services/order", "module.go")
	added, err := addEntryToModuleSlice(fs, modPath, "Providers", "NewOrderGRPC")
	if err != nil || !added {
		t.Fatalf("expected addEntryToModuleSlice to add NewOrderGRPC, added=%v err=%v", added, err)
	}
	rawAfterAdd, _ := fs.ReadFile(modPath)
	if !strings.Contains(string(rawAfterAdd), "NewOrderGRPC,") {
		t.Fatalf("expected NewOrderGRPC to be added:\n%s", string(rawAfterAdd))
	}

	removed, err := removeEntryFromModuleFile(fs, modPath, "GRPC,")
	if err != nil || !removed {
		t.Fatalf("expected removeEntryFromModuleFile to remove NewOrderGRPC, removed=%v err=%v", removed, err)
	}
	rawAfterRem, _ := fs.ReadFile(modPath)
	if strings.Contains(string(rawAfterRem), "NewOrderGRPC") {
		t.Fatalf("expected NewOrderGRPC to be removed:\n%s", string(rawAfterRem))
	}
}

func TestRemServiceUseCase_UpdatesKernel(t *testing.T) {
	_, fs, gen := setupTempProject(t)
	genUC := NewGenServiceUseCase(gen, fs)
	const modPath = "github.com/mmycin/goforge"
	if err := genUC.Run("user", modPath, nil); err != nil {
		t.Fatalf("GenServiceUseCase(user): %v", err)
	}
	if err := genUC.Run("product", modPath, nil); err != nil {
		t.Fatalf("GenServiceUseCase(product): %v", err)
	}

	remUC := NewRemServiceUseCase(gen, fs)
	if err := remUC.Run("product", modPath, nil); err != nil {
		t.Fatalf("RemServiceUseCase(product): %v", err)
	}

	if gen.Exists("internal/services/product") {
		t.Fatalf("expected internal/services/product to be deleted")
	}

	kernelRaw, err := fs.ReadFile("internal/services/kernel.go")
	if err != nil {
		t.Fatalf("read kernel.go: %v", err)
	}
	kernel := string(kernelRaw)
	if strings.Contains(kernel, "product") {
		t.Fatalf("expected kernel.go not to reference product:\n%s", kernel)
	}
	if !strings.Contains(kernel, "user.Module(),") {
		t.Fatalf("expected kernel.go to retain user.Module():\n%s", kernel)
	}
}

func TestRemGRPCUseCase_PatchesModulesAndBootKernel(t *testing.T) {
	_, fs, gen := setupTempProject(t)
	genUC := NewGenServiceUseCase(gen, fs)
	if err := genUC.Run("user", "github.com/mmycin/goforge", nil); err != nil {
		t.Fatalf("GenServiceUseCase(user): %v", err)
	}

	if _, err := addEntryToModuleSlice(fs, "internal/services/user/module.go", "Providers", "NewUserGRPC"); err != nil {
		t.Fatalf("addEntryToModuleSlice(NewUserGRPC): %v", err)
	}
	if err := fs.MkdirAll("boot", 0755); err != nil {
		t.Fatalf("mkdir boot: %v", err)
	}
	bootKernelSample := `package boot

func NewWithConfig(cfg *config.AllConfig, modules ...Module) (*App, error) {
	container := di.New(
		di.AutoCollect[server.Router](),
		di.AutoCollect[server.GRPCRegistry](),
		di.WithModules(allModules...),
	)
	routers := di.Collect[server.Router](container)
	grpcRegistries := di.Collect[server.GRPCRegistry](container)
	servers := buildServers(cfg, routers, grpcRegistries)
	return &App{Servers: servers}, nil
}

func NewApp(httpRouters []server.Router, grpcRegistries []server.GRPCRegistry) (*App, error) {
	servers := buildServers(cfg, httpRouters, grpcRegistries)
	return &App{Servers: servers}, nil
}

func buildServers(cfg *config.AllConfig, routers []server.Router, grpcRegistries []server.GRPCRegistry) []server.Server {
	return nil
}
`
	if err := fs.WriteFile("boot/kernel.go", []byte(bootKernelSample), 0644); err != nil {
		t.Fatalf("write boot/kernel.go: %v", err)
	}

	remGRPC := NewRemGRPCUseCase(fs, infra.NewOsExecutor())
	if err := remGRPC.patchServicesKernel(); err != nil {
		t.Fatalf("patchServicesKernel: %v", err)
	}
	if err := remGRPC.patchBootKernel(); err != nil {
		t.Fatalf("patchBootKernel: %v", err)
	}

	modRaw, _ := fs.ReadFile("internal/services/user/module.go")
	if strings.Contains(string(modRaw), "NewUserGRPC") {
		t.Fatalf("expected NewUserGRPC stripped from module.go:\n%s", string(modRaw))
	}

	bootRaw, _ := fs.ReadFile("boot/kernel.go")
	bootSrc := string(bootRaw)
	for _, unwanted := range []string{"GRPCRegistry", "grpcRegistries", "buildServers"} {
		if strings.Contains(bootSrc, unwanted) {
			t.Fatalf("expected %q stripped from boot/kernel.go:\n%s", unwanted, bootSrc)
		}
	}
}
