package setup_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/setup"
)

func TestDefaultModules(t *testing.T) {
	// Test that defaultModules returns expected modules
	modules := setup.DefaultModules()

	if len(modules) == 0 {
		t.Error("DefaultModules() should return at least one module")
	}

	// Verify it contains bpl-identifier
	found := false
	for _, m := range modules {
		if m == "bpl-identifier" {
			found = true
			break
		}
	}

	if !found {
		t.Error("DefaultModules() should contain 'bpl-identifier'")
	}
}

func TestBuildMessage_Unix(t *testing.T) {
	tests := []struct {
		name     string
		shell    string
		contains string
	}{
		{
			name:     "bash shell",
			shell:    "/bin/bash",
			contains: ".bashrc",
		},
		{
			name:     "zsh shell",
			shell:    "/bin/zsh",
			contains: ".zshrc",
		},
		{
			name:     "fish shell",
			shell:    "/usr/bin/fish",
			contains: "config.fish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SHELL", tt.shell)

			paths := []string{"/path/one", "/path/two"}
			message := setup.BuildMessage(paths)

			if message == "" {
				t.Error("BuildMessage() should return non-empty message")
			}

			if !contains(message, tt.contains) {
				t.Errorf("BuildMessage() for %s should contain %q", tt.shell, tt.contains)
			}
		})
	}
}

func TestBuildMessage_IncludesPaths(t *testing.T) {
	paths := []string{"/custom/path", "/another/path"}
	message := setup.BuildMessage(paths)

	if !contains(message, "/custom/path") {
		t.Error("BuildMessage() should include the provided paths")
	}
}

func TestCreatePaths(t *testing.T) {
	// Create a temp directory for BOSS_HOME
	tempDir := t.TempDir()
	t.Setenv("BOSS_HOME", tempDir)

	// Create boss home structure
	bossHome := filepath.Join(tempDir, consts.FolderBossHome)
	if err := os.MkdirAll(bossHome, 0755); err != nil {
		t.Fatalf("Failed to create boss home: %v", err)
	}

	// Call CreatePaths
	setup.CreatePaths()

	// Verify env/bpl was created
	envBplPath := filepath.Join(bossHome, consts.FolderEnvBpl)
	if _, err := os.Stat(envBplPath); os.IsNotExist(err) {
		t.Error("CreatePaths() should create env/bpl directory")
	}
}

func contains(s, substr string) bool {
	if len(s) == 0 || len(substr) == 0 {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestMigratorFunctions(t *testing.T) {
	// Test that migrator functions exist and are callable
	t.Run("DefaultModules returns correct format", func(t *testing.T) {
		modules := setup.DefaultModules()

		for _, module := range modules {
			if module == "" {
				t.Error("Module name should not be empty")
			}
		}
	})
}

func TestCreatePathsIdempotent(t *testing.T) {
	// Create a temp directory for BOSS_HOME
	tempDir := t.TempDir()
	t.Setenv("BOSS_HOME", tempDir)

	// Create boss home structure
	bossHome := filepath.Join(tempDir, consts.FolderBossHome)
	if err := os.MkdirAll(bossHome, 0755); err != nil {
		t.Fatalf("Failed to create boss home: %v", err)
	}

	// Call CreatePaths twice - should not panic
	setup.CreatePaths()
	setup.CreatePaths()
}
