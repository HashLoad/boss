package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install a new dependency",
	Long:    `This command install a new dependency`,
	Aliases: []string{"i", "add"},
	Run: func(cmd *cobra.Command, args []string) {
		core.InstallModules(args, true)
	},
}

func init() {
	RootCmd.AddCommand(installCmd)
}
