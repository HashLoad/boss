package cmd

import (
	"github.com/hashload/boss/internal/version"
	"github.com/hashload/boss/pkg/msg"
	"github.com/spf13/cobra"
)

func versionCmdRegister(root *cobra.Command) {
	var versionCmd = &cobra.Command{
		Use:     "version",
		Short:   "Show cli version",
		Long:    `This command show the client version`,
		Aliases: []string{"v"},
		Example: `  Print version:
  boss version`,
		Run: func(_ *cobra.Command, _ []string) {
			printVersion()
		},
	}

	root.AddCommand(versionCmd)
}

func printVersion() {
	v := version.Get()

	msg.Info("Boss CLI Version: %s", v.Version)
	msg.Info("Go Version: %s", v.GoVersion)
	msg.Info("Git Commit: %s", v.GitCommit)
}
