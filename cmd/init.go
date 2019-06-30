package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project",
	Long:  `This command initialize a new project`,
	Run: func(cmd *cobra.Command, args []string) {
		core.InitializeBossPackage()
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
