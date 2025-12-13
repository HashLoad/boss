package cli

import (
	"github.com/hashload/boss/internal/core/services/scripts"
	"github.com/spf13/cobra"
)

// runCmdRegister registers the run command
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
