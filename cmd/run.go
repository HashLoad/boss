package cmd

import (
	"github.com/hashload/boss/pkg/scripts"
	"github.com/spf13/cobra"
)

func runCmdRegister(root *cobra.Command) {
	var runScript = &cobra.Command{
		Use:   "run",
		Short: "Run cmd script",
		Long:  `Run cmd script`,
		Run: func(_ *cobra.Command, args []string) {
			scripts.Run(args)
		},
	}

	root.AddCommand(runScript)
}
