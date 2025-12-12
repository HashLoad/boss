package env_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
)

func TestLoadConfiguration_NewConfig(t *testing.T) {
	tempDir := t.TempDir()

	config, err := env.LoadConfiguration(tempDir)

	// Should return error for non-existent file, but still return a config
	if err == nil {
		t.Log("LoadConfiguration() returned nil error (file may exist)")
	}

	if config == nil {
		t.Fatal("LoadConfiguration() should return a configuration even on error")
	}

	// Default values should be set
	if config.PurgeTime != 3 {
		t.Errorf("PurgeTime = %d, want 3", config.PurgeTime)
	}

	if config.Auth == nil {
		t.Error("Auth should not be nil")
	}
}

func TestLoadConfiguration_ExistingConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Create a valid config file
	configData := map[string]any{
		"id":                    "test-key",
		"purge_after":           7,
		"internal_refresh_rate": 10,
		"git_embedded":          false,
		"auth":                  map[string]any{},
	}
	data, _ := json.Marshal(configData)

	configPath := filepath.Join(tempDir, consts.BossConfigFile)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := env.LoadConfiguration(tempDir)

	if err != nil {
		t.Errorf("LoadConfiguration() error = %v", err)
	}

	if config == nil {
		t.Fatal("LoadConfiguration() should return a configuration")
	}

	if config.PurgeTime != 7 {
		t.Errorf("PurgeTime = %d, want 7", config.PurgeTime)
	}

	if config.InternalRefreshRate != 10 {
		t.Errorf("InternalRefreshRate = %d, want 10", config.InternalRefreshRate)
	}

	if config.GitEmbedded != false {
		t.Error("GitEmbedded should be false")
	}
}

func TestLoadConfiguration_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()

	// Create an invalid JSON file
	configPath := filepath.Join(tempDir, consts.BossConfigFile)
	if err := os.WriteFile(configPath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := env.LoadConfiguration(tempDir)

	// Should return error but still return a default config
	if err == nil {
		t.Error("LoadConfiguration() should return error for invalid JSON")
	}

	if config == nil {
		t.Fatal("LoadConfiguration() should return a default configuration on error")
	}

	// Should have default values
	if config.PurgeTime != 3 {
		t.Errorf("PurgeTime = %d, want default 3", config.PurgeTime)
	}
}

func TestConfiguration_SaveConfiguration(t *testing.T) {
	tempDir := t.TempDir()

	// Load a new configuration
	config, _ := env.LoadConfiguration(tempDir)

	// Modify it
	config.PurgeTime = 5
	config.InternalRefreshRate = 15

	// Save it
	config.SaveConfiguration()

	// Verify the file was created
	configPath := filepath.Join(tempDir, consts.BossConfigFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("SaveConfiguration() should create config file")
	}

	// Load it again and verify
	loaded, err := env.LoadConfiguration(tempDir)
	if err != nil {
		t.Errorf("Failed to load saved configuration: %v", err)
	}

	if loaded.PurgeTime != 5 {
		t.Errorf("Loaded PurgeTime = %d, want 5", loaded.PurgeTime)
	}

	if loaded.InternalRefreshRate != 15 {
		t.Errorf("Loaded InternalRefreshRate = %d, want 15", loaded.InternalRefreshRate)
	}
}

func TestConfiguration_GetAuth_Nil(t *testing.T) {
	tempDir := t.TempDir()

	config, _ := env.LoadConfiguration(tempDir)

	// GetAuth for non-existent repo should return nil
	auth := config.GetAuth("nonexistent-repo")

	if auth != nil {
		t.Error("GetAuth() for non-existent repo should return nil")
	}
}

func TestAuth_SetAndGetUser(t *testing.T) {
	tempDir := t.TempDir()

	config, _ := env.LoadConfiguration(tempDir)

	// Create a new auth entry
	config.Auth["github.com"] = &env.Auth{}
	config.Auth["github.com"].SetUser("testuser")

	// Get the user back
	user := config.Auth["github.com"].GetUser()

	if user != "testuser" {
		t.Errorf("GetUser() = %q, want %q", user, "testuser")
	}
}

func TestAuth_SetAndGetPassword(t *testing.T) {
	tempDir := t.TempDir()

	config, _ := env.LoadConfiguration(tempDir)

	// Create a new auth entry
	config.Auth["github.com"] = &env.Auth{}
	config.Auth["github.com"].SetPass("testpass")

	// Get the password back
	pass := config.Auth["github.com"].GetPassword()

	if pass != "testpass" {
		t.Errorf("GetPassword() = %q, want %q", pass, "testpass")
	}
}

func TestAuth_SetAndGetPassPhrase(t *testing.T) {
	tempDir := t.TempDir()

	config, _ := env.LoadConfiguration(tempDir)

	// Create a new auth entry
	config.Auth["github.com"] = &env.Auth{}
	config.Auth["github.com"].SetPassPhrase("testphrase")

	// Get the passphrase back
	phrase := config.Auth["github.com"].GetPassPhrase()

	if phrase != "testphrase" {
		t.Errorf("GetPassPhrase() = %q, want %q", phrase, "testphrase")
	}
}

func TestAuth_UseSSH_Flag(t *testing.T) {
	auth := &env.Auth{
		UseSSH: true,
		Path:   "/path/to/key",
	}

	if !auth.UseSSH {
		t.Error("UseSSH should be true")
	}

	if auth.Path != "/path/to/key" {
		t.Errorf("Path = %q, want %q", auth.Path, "/path/to/key")
	}
}

func TestConfiguration_GetAuth_BasicAuth(t *testing.T) {
	tempDir := t.TempDir()

	config, _ := env.LoadConfiguration(tempDir)

	// Create auth entry with basic auth (UseSSH = false)
	config.Auth["github.com"] = &env.Auth{
		UseSSH: false,
	}
	config.Auth["github.com"].SetUser("user")
	config.Auth["github.com"].SetPass("pass")

	// GetAuth should return BasicAuth
	auth := config.GetAuth("github.com")

	if auth == nil {
		t.Error("GetAuth() should return auth method for existing repo")
	}

	// Type should be BasicAuth
	if auth.Name() != "http-basic-auth" {
		t.Errorf("Auth type = %q, want http-basic-auth", auth.Name())
	}
}
