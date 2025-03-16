package cmd

import (
	"github.com/hashload/boss/pkg/installer"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update dependencies",
	Long:    `This command update installed dependencies`,
	Aliases: []string{"up"},
	Run: func(cmd *cobra.Command, args []string) {
		installer.InstallModules(args, false, false)
	},
}

func init() {
	root.AddCommand(updateCmd)
}
