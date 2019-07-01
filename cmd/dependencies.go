package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var dependenciesCmd = &cobra.Command{
	Use:     "dependencies",
	Short:   "Print all dependencies",
	Long:    `This command print all dependencies and your versions`,
	Aliases: []string{"dep"},
	Run: func(cmd *cobra.Command, args []string) {
		core.PrintDependencies()
	},
}

func init() {
	RootCmd.AddCommand(dependenciesCmd)
}
