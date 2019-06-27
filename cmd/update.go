package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "update dependencies",
	Long:    `update dependencies`,
	Aliases: []string{"up"},
	Run: func(cmd *cobra.Command, args []string) {
		core.InstallModules(args, false)
	},
}

func init() {
	RootCmd.AddCommand(updateCmd)
}
