package config

import (
	"github.com/spf13/cobra"
)

var CmdConfig = &cobra.Command{
	Use:   "config",
	Short: "Configurations",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func InitializeConfig(root *cobra.Command) {
	root.AddCommand(CmdConfig)
}
