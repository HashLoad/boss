package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a dependency",
	Long:  `Remove a dependency`,
	Run: func(cmd *cobra.Command, args []string) {
		core.RemoveModules(args)
	},
}

func init() {
	RootCmd.AddCommand(removeCmd)
}
