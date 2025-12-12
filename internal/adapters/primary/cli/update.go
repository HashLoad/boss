package cli

import (
	"github.com/hashload/boss/internal/core/services/installer"
	"github.com/spf13/cobra"
)

func updateCmdRegister(root *cobra.Command) {
	var updateCmd = &cobra.Command{
		Use:     "update",
		Short:   "Update dependencies",
		Long:    `This command update installed dependencies`,
		Aliases: []string{"up"},
		Run: func(_ *cobra.Command, args []string) {
			installer.InstallModules(args, false, false)
		},
	}

	root.AddCommand(updateCmd)
}
