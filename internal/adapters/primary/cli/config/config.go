// Package config provides Boss configuration management commands.
package config

import (
	"github.com/spf13/cobra"
)

const cmdNameConfig = "config"

// RegisterConfigCommand registers the config command.
func RegisterConfigCommand(root *cobra.Command) {
	configCmd := &cobra.Command{
		Use:   cmdNameConfig,
		Short: "Configurations",
	}

	root.AddCommand(configCmd)
	delphiCmd(configCmd)
	registryGitCmd(configCmd)
	RegisterCmd(configCmd)
}
