package cmd

import (
	"github.com/hashload/boss/models"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a dependency",
	Long:  `Remove a dependency`,
	Run: func(cmd *cobra.Command, args []string) {
		loadPackage, e := models.LoadPackage(false)
		if e != nil {
			e.Error()
		}

		for e := range args {
			loadPackage.RemoveDependency(args[e])
		}
		loadPackage.Save()
	},
}

func init() {
	RootCmd.AddCommand(removeCmd)
}
