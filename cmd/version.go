package cmd

import (
	"fmt"

	"github.com/hashload/boss/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Show cli version",
	Long:    `This command show the client version`,
	Aliases: []string{"v"},
	Example: `  Print version:
  boss version`,
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

func printVersion() {
	v := version.Get()
	fmt.Println("Version        ", v.Version)
	fmt.Println("Git commit     ", v.GitCommit)
	fmt.Println("Go version     ", v.GoVersion)
}
