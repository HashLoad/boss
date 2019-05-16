package cmd

import (
	"github.com/hashload/boss/core/gc"
	"github.com/hashload/boss/core/installer"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/spf13/cobra"
	"os"
)

var installCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install a dependency",
	Long:    `Install a dependency`,
	Aliases: []string{"i"},
	Run: func(cmd *cobra.Command, args []string) {
		pkg, e := models.LoadPackage(env.Global)

		if e != nil {
			if os.IsNotExist(e) {
				msg.Die("boss.json not exists in " + env.GetCurrentDir())
			} else {
				msg.Die("Fail on open dependencies file: %s", e)
			}
		}

		if env.Global {
			installer.GlobalInstall(args, pkg)
		} else {
			installer.LocalInstall(args, pkg)
		}

		gc.RunGC()
	},
}

func init() {
	RootCmd.AddCommand(installCmd)
}
