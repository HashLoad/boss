package config

import (
	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configurations",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func InitilizeConfig(root *cobra.Command) {
	root.AddCommand(ConfigCmd)
}
