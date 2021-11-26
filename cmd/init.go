package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var quiet bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project",
	Long:  "Initialize a new project and creates a boss.json file",
	Example: `  Initialize a new project:
  boss init

  Initialize a new project without having it ask any questions:
  boss init --quiet`,
	Run: func(cmd *cobra.Command, args []string) {
		core.InitializeBossPackage(quiet)
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "without asking questions")
}
