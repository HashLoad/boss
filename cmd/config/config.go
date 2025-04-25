package config

import (
	"github.com/spf13/cobra"
)

func RegisterConfigCommand(root *cobra.Command) {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configurations",
	}

	root.AddCommand(configCmd)
	delphiCmd(configCmd)
}
