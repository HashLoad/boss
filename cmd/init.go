package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project",
	Long:  `Initialize a new project`,
	Run: func(cmd *cobra.Command, args []string) {
		core.InitilizeBossPackage()
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
