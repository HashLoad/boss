//nolint:testpackage // Testing internal command registration
package config

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestRegisterConfigCommand tests config command registration.
func TestRegisterConfigCommand(t *testing.T) {
	root := &cobra.Command{Use: "boss"}

	RegisterConfigCommand(root)

	// Find config command
	var configCmd *cobra.Command
	for _, cmd := range root.Commands() {
		if cmd.Use == "config" {
			configCmd = cmd
			break
		}
	}

	if configCmd == nil {
		t.Fatal("Config command not found")
	}

	if configCmd.Short == "" {
		t.Error("Config command should have a short description")
	}

	// Check subcommands exist
	subcommands := configCmd.Commands()
	if len(subcommands) == 0 {
		t.Error("Config command should have subcommands")
	}
}

// TestConfigSubcommands tests config subcommand structure.
func TestConfigSubcommands(t *testing.T) {
	root := &cobra.Command{Use: "boss"}
	RegisterConfigCommand(root)

	var configCmd *cobra.Command
	for _, cmd := range root.Commands() {
		if cmd.Use == "config" {
			configCmd = cmd
			break
		}
	}

	if configCmd == nil {
		t.Fatal("Config command not found")
	}

	expectedSubcommands := []string{"delphi", "git"}
	foundSubcommands := make(map[string]bool)

	for _, cmd := range configCmd.Commands() {
		foundSubcommands[cmd.Use] = true
	}

	for _, expected := range expectedSubcommands {
		if !foundSubcommands[expected] {
			t.Errorf("Expected subcommand '%s' not found", expected)
		}
	}
}
