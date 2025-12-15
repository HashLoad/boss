// Package config provides Boss configuration management commands.
package config

import (
	"github.com/spf13/cobra"
)

// RegisterConfigCommand registers the config command.
func RegisterConfigCommand(root *cobra.Command) {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configurations",
	}

	root.AddCommand(configCmd)
	delphiCmd(configCmd)
	registryGitCmd(configCmd)
	RegisterCmd(configCmd)
}
