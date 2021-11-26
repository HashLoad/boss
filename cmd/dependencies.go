package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var showVersion bool

var dependenciesCmd = &cobra.Command{
	Use:     "dependencies",
	Short:   "Print all project dependencies",
	Long:    "Print all project dependencies with or without version control",
	Aliases: []string{"dep", "ls", "list", "ll", "la"},
	Example: `  Listing all dependencies:
  boss dependencies

  Listing all dependencies with version control:
  boss dependencies --version

  List package dependencies:
  boss dependencies <pkg>

  List package dependencies with version control:
  boss dependencies <pkg> --version`,
	Run: func(cmd *cobra.Command, args []string) {
		core.PrintDependencies(showVersion)
	},
}

func init() {
	RootCmd.AddCommand(dependenciesCmd)
	dependenciesCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show dependency version")
}
