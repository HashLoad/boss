package env_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
)

func TestSetGlobal_GetGlobal(t *testing.T) {
	// Save original state
	original := env.GetGlobal()
	defer env.SetGlobal(original)

	tests := []struct {
		name     string
		setValue bool
	}{
		{"set global true", true},
		{"set global false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env.SetGlobal(tt.setValue)
			if got := env.GetGlobal(); got != tt.setValue {
				t.Errorf("GetGlobal() = %v, want %v", got, tt.setValue)
			}
		})
	}
}

func TestSetInternal_GetInternal(t *testing.T) {
	// Save original state
	original := env.GetInternal()
	defer env.SetInternal(original)

	tests := []struct {
		name     string
		setValue bool
	}{
		{"set internal true", true},
		{"set internal false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env.SetInternal(tt.setValue)
			if got := env.GetInternal(); got != tt.setValue {
				t.Errorf("GetInternal() = %v, want %v", got, tt.setValue)
			}
		})
	}
}

func TestGlobalConfiguration(t *testing.T) {
	config := env.GlobalConfiguration()
	// GlobalConfiguration should never return nil
	if config == nil {
		t.Error("GlobalConfiguration() should not return nil")
	}
}

func TestGetBossHome(t *testing.T) {
	t.Run("with BOSS_HOME set", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("BOSS_HOME", tempDir)

		result := env.GetBossHome()
		expected := filepath.Join(tempDir, consts.FolderBossHome)
		if result != expected {
			t.Errorf("GetBossHome() = %q, want %q", result, expected)
		}
	})

	t.Run("without BOSS_HOME", func(t *testing.T) {
		// Note: cannot unset env in parallel tests, just verify the function works
		result := env.GetBossHome()
		// Should contain the boss home folder
		if !strings.HasSuffix(result, consts.FolderBossHome) {
			t.Errorf("GetBossHome() = %q, should end with %q", result, consts.FolderBossHome)
		}
	})
}

func TestGetCacheDir(t *testing.T) {
	result := env.GetCacheDir()

	// Should contain "cache" and be under boss home
	if !strings.Contains(result, "cache") {
		t.Errorf("GetCacheDir() = %q, should contain 'cache'", result)
	}
}

func TestGetBossFile(t *testing.T) {
	result := env.GetBossFile()

	// Should end with boss.json
	if !strings.HasSuffix(result, consts.FilePackage) {
		t.Errorf("GetBossFile() = %q, should end with %q", result, consts.FilePackage)
	}
}

func TestGetModulesDir(t *testing.T) {
	result := env.GetModulesDir()

	// Should end with the dependencies folder
	if !strings.HasSuffix(result, consts.FolderDependencies) {
		t.Errorf("GetModulesDir() = %q, should end with %q", result, consts.FolderDependencies)
	}
}

func TestGetCurrentDir(t *testing.T) {
	// Save original global state
	originalGlobal := env.GetGlobal()
	defer env.SetGlobal(originalGlobal)

	t.Run("when not global", func(t *testing.T) {
		env.SetGlobal(false)
		result := env.GetCurrentDir()

		// Should be current working directory
		cwd, _ := os.Getwd()
		if result != cwd {
			t.Errorf("GetCurrentDir() = %q, want %q", result, cwd)
		}
	})

	t.Run("when global", func(t *testing.T) {
		env.SetGlobal(true)
		result := env.GetCurrentDir()

		// Should be under boss home with dependencies folder
		bossHome := env.GetBossHome()
		if !strings.HasPrefix(result, bossHome) {
			t.Errorf("GetCurrentDir() = %q, should be under boss home %q", result, bossHome)
		}
	})
}

func TestGetGlobalEnvPaths(t *testing.T) {
	bossHome := env.GetBossHome()

	t.Run("GetGlobalEnvBpl", func(t *testing.T) {
		result := env.GetGlobalEnvBpl()
		if !strings.HasPrefix(result, bossHome) {
			t.Errorf("GetGlobalEnvBpl() = %q, should be under boss home", result)
		}
		if !strings.Contains(result, consts.FolderEnvBpl) {
			t.Errorf("GetGlobalEnvBpl() = %q, should contain %q", result, consts.FolderEnvBpl)
		}
	})

	t.Run("GetGlobalEnvDcp", func(t *testing.T) {
		result := env.GetGlobalEnvDcp()
		if !strings.HasPrefix(result, bossHome) {
			t.Errorf("GetGlobalEnvDcp() = %q, should be under boss home", result)
		}
		if !strings.Contains(result, consts.FolderEnvDcp) {
			t.Errorf("GetGlobalEnvDcp() = %q, should contain %q", result, consts.FolderEnvDcp)
		}
	})

	t.Run("GetGlobalEnvDcu", func(t *testing.T) {
		result := env.GetGlobalEnvDcu()
		if !strings.HasPrefix(result, bossHome) {
			t.Errorf("GetGlobalEnvDcu() = %q, should be under boss home", result)
		}
		if !strings.Contains(result, consts.FolderEnvDcu) {
			t.Errorf("GetGlobalEnvDcu() = %q, should contain %q", result, consts.FolderEnvDcu)
		}
	})
}

func TestGetGlobalBinPath(t *testing.T) {
	result := env.GetGlobalBinPath()
	bossHome := env.GetBossHome()

	if !strings.HasPrefix(result, bossHome) {
		t.Errorf("GetGlobalBinPath() = %q, should be under boss home", result)
	}
	if !strings.Contains(result, consts.BinFolder) {
		t.Errorf("GetGlobalBinPath() = %q, should contain %q", result, consts.BinFolder)
	}
}

func TestHashDelphiPath(t *testing.T) {
	// Save original state
	originalInternal := env.GetInternal()
	defer env.SetInternal(originalInternal)

	t.Run("not internal", func(t *testing.T) {
		env.SetInternal(false)
		result := env.HashDelphiPath()

		// Should be a 32-character hex string (MD5)
		if len(result) != 32 {
			t.Errorf("HashDelphiPath() length = %d, want 32", len(result))
		}
	})

	t.Run("internal", func(t *testing.T) {
		env.SetInternal(true)
		result := env.HashDelphiPath()

		// Should contain the internal dir prefix
		if !strings.HasPrefix(result, consts.BossInternalDir) {
			t.Errorf("HashDelphiPath() = %q, should have internal prefix %q", result, consts.BossInternalDir)
		}
	})
}

func TestGetInternalGlobalDir(t *testing.T) {
	// Save original state
	originalInternal := env.GetInternal()
	defer env.SetInternal(originalInternal)

	// Reset to known state
	env.SetInternal(false)

	result := env.GetInternalGlobalDir()

	// Should be under boss home
	bossHome := env.GetBossHome()
	if !strings.HasPrefix(result, bossHome) {
		t.Errorf("GetInternalGlobalDir() = %q, should be under boss home", result)
	}

	// Should contain dependencies folder
	if !strings.Contains(result, consts.FolderDependencies) {
		t.Errorf("GetInternalGlobalDir() = %q, should contain dependencies folder", result)
	}

	// Original internal state should be preserved
	if env.GetInternal() != false {
		t.Error("GetInternalGlobalDir() should preserve original internal state")
	}
}

func TestGetDcc32Dir(_ *testing.T) {
	// This function depends on system configuration
	// We just verify it doesn't panic
	result := env.GetDcc32Dir()
	_ = result // May be empty string if Delphi is not installed
}
