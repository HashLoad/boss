package cmd

import (
	"github.com/hashload/boss/consts"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "show cli version",
	Aliases: []string{"v"},
	Run: func(cmd *cobra.Command, args []string) {
		println(consts.Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
