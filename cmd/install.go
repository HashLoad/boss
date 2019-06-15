package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install a dependency",
	Long:    `Install a dependency`,
	Aliases: []string{"i"},
	Run: func(cmd *cobra.Command, args []string) {
		core.InstallModules(args, true)
	},
}

func init() {
	RootCmd.AddCommand(installCmd)
}
