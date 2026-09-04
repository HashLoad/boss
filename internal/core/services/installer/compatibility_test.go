//nolint:testpackage // Testing internal implementation details
package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashload/boss/internal/adapters/secondary/filesystem"
	"github.com/hashload/boss/internal/adapters/secondary/repository"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/packages"
	"github.com/hashload/boss/pkg/pkgmanager"
)

// installPackageManager wires the real filesystem-backed package service so the
// compatibility check reads the module directory the test just created.
func installPackageManager(t *testing.T) {
	t.Helper()

	fs := filesystem.NewOSFileSystem()
	packageRepo := repository.NewFilePackageRepository(fs)
	lockRepo := repository.NewFileLockRepository(fs)
	pkgmanager.SetInstance(packages.NewPackageService(packageRepo, lockRepo))
}

// newCompatibilityContext returns a context targeting modulesDir with a Win32 toolchain.
func newCompatibilityContext(modulesDir string) *installContext {
	return &installContext{
		modulesDir: modulesDir,
		root: &domain.Package{
			Toolchain: &domain.PackageToolchain{Platform: "Win32"},
		},
		options: InstallOptions{},
	}
}

// writeModule creates modulesDir/<dep name> and, when content is not empty, its boss.json.
func writeModule(t *testing.T, modulesDir string, dep domain.Dependency, content string) {
	t.Helper()

	depPath := filepath.Join(modulesDir, dep.Name())
	if err := os.MkdirAll(depPath, 0755); err != nil {
		t.Fatalf("failed to create module dir: %v", err)
	}
	if content == "" {
		return
	}
	if err := os.WriteFile(filepath.Join(depPath, "boss.json"), []byte(content), 0600); err != nil {
		t.Fatalf("failed to write boss.json: %v", err)
	}
}

// TestVerifyDependencyCompatibility_WithoutBossJSON covers plain source libraries
// that carry no boss.json at all - github.com/academiadocodigo/localcache4d and
// github.com/andre-djsystem/hashlib4pascal are two, and Kastri and Alcinoe are more.
// They declare no engines, so there is no platform constraint to violate and the
// install must go through.
func TestVerifyDependencyCompatibility_WithoutBossJSON(t *testing.T) {
	installPackageManager(t)

	modulesDir := t.TempDir()
	dep := domain.Dependency{Repository: "github.com/andre-djsystem/hashlib4pascal"}
	writeModule(t, modulesDir, dep, "")

	warning, err := newCompatibilityContext(modulesDir).verifyDependencyCompatibility(dep)
	if err != nil {
		t.Fatalf("a dependency without boss.json must not fail the install, got: %v", err)
	}
	if warning != "" {
		t.Errorf("expected no warning, got %q", warning)
	}
}

// TestVerifyDependencyCompatibility_MalformedBossJSON guards the other side of the
// fix: a boss.json that exists but cannot be parsed is a real defect and must still
// be reported instead of being silently ignored.
func TestVerifyDependencyCompatibility_MalformedBossJSON(t *testing.T) {
	installPackageManager(t)

	modulesDir := t.TempDir()
	dep := domain.Dependency{Repository: "github.com/example/broken"}
	writeModule(t, modulesDir, dep, "{ this is not json")

	if _, err := newCompatibilityContext(modulesDir).verifyDependencyCompatibility(dep); err == nil {
		t.Fatal("expected a malformed boss.json to be reported, got nil")
	}
}

// TestVerifyDependencyCompatibility_UnsupportedPlatform proves the check itself keeps
// working: a dependency that declares engines without the target platform is refused.
func TestVerifyDependencyCompatibility_UnsupportedPlatform(t *testing.T) {
	installPackageManager(t)

	modulesDir := t.TempDir()
	dep := domain.Dependency{Repository: "github.com/example/linuxonly"}
	writeModule(t, modulesDir, dep, `{"name":"linuxonly","engines":{"platforms":["Linux64"]}}`)

	ctx := newCompatibilityContext(modulesDir)
	ctx.options.Strict = true

	if _, err := ctx.verifyDependencyCompatibility(dep); err == nil {
		t.Fatal("expected an unsupported platform to be refused in strict mode")
	}
}
