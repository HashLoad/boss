package cmd

import (
	"github.com/hashload/boss/core/scripts"
	"github.com/spf13/cobra"
)

var runScript = &cobra.Command{
	Use:   "run",
	Short: "Run cmd script",
	Long:  `Run cmd script`,
	Run: func(cmd *cobra.Command, args []string) {
		scripts.Run(args)
	},
}

func init() {
	RootCmd.AddCommand(runScript)
}
