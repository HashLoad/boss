package cmd

import (
	"os"

	"github.com/hashload/boss/cmd/config"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/gc"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/setup"
	"github.com/hashload/boss/utils"

	"github.com/spf13/cobra"
)

var versionPrint bool

var root = &cobra.Command{
	Use:   "boss",
	Short: "Dependency Manager for Delphi",
	Long:  "Dependency Manager for Delphi",
	RunE: func(cmd *cobra.Command, args []string) error {
		if versionPrint {
			printVersion()
		} else {
			return cmd.Help()
		}
		return nil
	},
}

func Execute() {
	root.PersistentFlags().BoolVarP(&env.Global, "global", "g", false, "global environment")
	root.PersistentFlags().BoolVarP(&msg.DebugEnable, "debug", "d", false, "debug")
	root.Flags().BoolVarP(&versionPrint, "version", "v", false, "show cli version")

	msg.DebugEnable = utils.Contains(os.Args, "-d")

	setup.Initialize()

	config.InitializeConfig(root)

	gc.RunGC(false)

	config.RegisterCmd(root)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
