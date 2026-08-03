package env_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/utils/crypto"
)

// An SSH key without a passphrase stores nothing in keypass. Decrypting that
// empty value used to fail with "cipher text block size is too short" and, since
// the getters call msg.Die, took the whole process down — every `boss install`
// and `boss update` on a machine that had ever run `boss login -s` exited 1.
func TestAuthGettersTreatEmptyValuesAsUnset(t *testing.T) {
	auth := &env.Auth{UseSSH: true, Path: "/home/user/.ssh/id_ed25519"}

	if got := auth.GetPassPhrase(); got != "" {
		t.Errorf("GetPassPhrase() = %q, want %q", got, "")
	}
	if got := auth.GetUser(); got != "" {
		t.Errorf("GetUser() = %q, want %q", got, "")
	}
	if got := auth.GetPassword(); got != "" {
		t.Errorf("GetPassword() = %q, want %q", got, "")
	}
}

func TestAuthRoundTripsEncryptedValues(t *testing.T) {
	auth := &env.Auth{}
	auth.SetUser("octocat")
	auth.SetPass("s3cr3t")
	auth.SetPassPhrase("phrase")

	if got := auth.GetUser(); got != "octocat" {
		t.Errorf("GetUser() = %q, want %q", got, "octocat")
	}
	if got := auth.GetPassword(); got != "s3cr3t" {
		t.Errorf("GetPassword() = %q, want %q", got, "s3cr3t")
	}
	if got := auth.GetPassPhrase(); got != "phrase" {
		t.Errorf("GetPassPhrase() = %q, want %q", got, "phrase")
	}
}

// Configurations written by Boss up to v3.0.12 keep the credentials under
// x/y/z. Loading and saving one has to round-trip those fields: dropping them
// destroys the credential before migration 7 ever gets a chance to convert it.
func TestLegacyAuthFieldsSurviveLoadAndSave(t *testing.T) {
	tempDir := t.TempDir()

	// LoadConfiguration wipes the auth map when the stored id does not match the
	// machine, so the fixture has to claim this machine.
	legacy := map[string]any{
		"id": crypto.Md5MachineID(),
		"auth": map[string]any{
			"github.com": map[string]any{
				"use":  true,
				"path": "/home/user/.ssh/id_ed25519",
				"x":    "bGVnYWN5LXVzZXI=",
				"y":    "bGVnYWN5LXBhc3M=",
				"z":    "bGVnYWN5LXBocmFzZQ==",
			},
		},
		"config_version": 6,
	}
	data, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("Failed to marshal legacy config: %v", err)
	}

	configPath := filepath.Join(tempDir, consts.BossConfigFile)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := env.LoadConfiguration(tempDir)
	if err != nil {
		t.Fatalf("LoadConfiguration() error = %v", err)
	}

	auth := config.Auth["github.com"]
	if auth == nil {
		t.Fatal("auth entry for github.com is missing")
	}
	if auth.LegacyUser != "bGVnYWN5LXVzZXI=" {
		t.Errorf("LegacyUser = %q, want the value read from x", auth.LegacyUser)
	}
	if auth.LegacyPass != "bGVnYWN5LXBhc3M=" {
		t.Errorf("LegacyPass = %q, want the value read from y", auth.LegacyPass)
	}
	if auth.LegacyPassPhrase != "bGVnYWN5LXBocmFzZQ==" {
		t.Errorf("LegacyPassPhrase = %q, want the value read from z", auth.LegacyPassPhrase)
	}

	config.SaveConfiguration()

	saved, err := os.ReadFile(configPath) // #nosec G304 -- test-owned temporary path
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	var reloaded map[string]any
	if err := json.Unmarshal(saved, &reloaded); err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	authMap, ok := reloaded["auth"].(map[string]any)
	if !ok {
		t.Fatal("saved config has no auth map")
	}
	entry, ok := authMap["github.com"].(map[string]any)
	if !ok {
		t.Fatal("saved config lost the github.com auth entry")
	}
	for _, field := range []string{"x", "y", "z"} {
		if _, found := entry[field]; !found {
			t.Errorf("saved config dropped the legacy field %q", field)
		}
	}
}
