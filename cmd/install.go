package cmd

import (
	"github.com/hashload/boss/core/gc"
	"github.com/hashload/boss/core/installer"
	"github.com/hashload/boss/env"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install a dependency",
	Long:    `Install a dependency`,
	Aliases: []string{"i"},
	Run: func(cmd *cobra.Command, args []string) {
		if env.Global {
			installer.GlobalInstall(args)
		} else {
			installer.LocalInstall(args)
		}

		gc.RunGC()
	},
}

func init() {
	RootCmd.AddCommand(installCmd)
}
