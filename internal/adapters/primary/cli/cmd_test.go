//nolint:testpackage // Testing internal command registration
package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

// TestRootCommand tests the root command structure.
func TestRootCommand(t *testing.T) {
	// We can't directly test Execute() as it calls os.Exit
	// But we can test the command registration

	// Create a mock root command to test command registration
	root := &cobra.Command{
		Use:   "boss",
		Short: "Dependency Manager for Delphi",
	}

	// Test that commands can be registered without panic
	t.Run("register commands", func(t *testing.T) {
		// These should not panic
		versionCmdRegister(root)

		// Verify command was added
		if root.Commands() == nil {
			t.Error("Expected commands to be registered")
		}
	})
}

// TestVersionCommand tests the version command.
func TestVersionCommand(t *testing.T) {
	root := &cobra.Command{Use: "boss"}
	versionCmdRegister(root)

	// Find the version command
	var versionCmd *cobra.Command
	for _, cmd := range root.Commands() {
		if cmd.Use == "version" {
			versionCmd = cmd
			break
		}
	}

	if versionCmd == nil {
		t.Fatal("Version command not found")
	}

	// Test command properties
	if versionCmd.Short == "" {
		t.Error("Version command should have a short description")
	}

	// Test aliases
	if len(versionCmd.Aliases) == 0 {
		t.Error("Version command should have aliases")
	}

	hasVAlias := false
	for _, alias := range versionCmd.Aliases {
		if alias == "v" {
			hasVAlias = true
			break
		}
	}
	if !hasVAlias {
		t.Error("Version command should have 'v' alias")
	}
}

// TestInstallCommand tests the install command registration.
func TestInstallCommand(t *testing.T) {
	root := &cobra.Command{Use: "boss"}
	installCmdRegister(root)

	// Find the install command
	var installCmd *cobra.Command
	for _, cmd := range root.Commands() {
		if cmd.Use == "install" {
			installCmd = cmd
			break
		}
	}

	if installCmd == nil {
		t.Fatal("Install command not found")
	}

	// Test aliases
	expectedAliases := map[string]bool{"i": false, "add": false}
	for _, alias := range installCmd.Aliases {
		if _, ok := expectedAliases[alias]; ok {
			expectedAliases[alias] = true
		}
	}

	for alias, found := range expectedAliases {
		if !found {
			t.Errorf("Install command should have '%s' alias", alias)
		}
	}

	// Test flags
	noSaveFlag := installCmd.Flags().Lookup("no-save")
	if noSaveFlag == nil {
		t.Error("Install command should have --no-save flag")
	}
}

// TestCommandHelp tests that commands have proper help text.
func TestCommandHelp(t *testing.T) {
	root := &cobra.Command{Use: "boss"}

	// Register all commands
	versionCmdRegister(root)
	installCmdRegister(root)

	for _, cmd := range root.Commands() {
		t.Run(cmd.Use, func(t *testing.T) {
			if cmd.Short == "" {
				t.Errorf("Command %s should have a short description", cmd.Use)
			}
			if cmd.Long == "" {
				t.Errorf("Command %s should have a long description", cmd.Use)
			}
		})
	}
}

// TestCommandOutput captures command output for testing.
func captureOutput(cmd *cobra.Command, args []string) (string, error) {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

// TestRootHelp tests that root command shows help.
func TestRootHelp(t *testing.T) {
	root := &cobra.Command{
		Use:   "boss",
		Short: "Dependency Manager for Delphi",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	output, err := captureOutput(root, []string{})
	if err != nil {
		t.Errorf("Root command should not error: %v", err)
	}

	if output == "" {
		t.Error("Root command should produce help output")
	}
}
