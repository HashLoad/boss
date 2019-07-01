package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Short:   "Uninstall a dependency",
	Long:    `This command uninstall a dependency`,
	Aliases: []string{"remove", "rm", "r", "un", "unlink"},
	Run: func(cmd *cobra.Command, args []string) {
		core.UninstallModules(args)
	},
}

func init() {
	RootCmd.AddCommand(uninstallCmd)
}
