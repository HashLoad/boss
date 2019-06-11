package cmd

import (
	"github.com/hashload/boss/core/installer"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a dependency",
	Long:  `Remove a dependency`,
	Run: func(cmd *cobra.Command, args []string) {
		pkg, e := models.LoadPackage(false)
		if e != nil {
			e.Error()
		}

		if pkg == nil {
			return
		}

		for e := range args {
			pkg.RemoveDependency(installer.ParseDependency(installer.ParseDependency(args[e])))
		}
		pkg.Save()

		if env.Global {
			installer.GlobalInstall([]string{}, pkg)
		} else {
			installer.LocalInstall([]string{}, pkg)
		}
	},
}

func init() {
	RootCmd.AddCommand(removeCmd)
}
