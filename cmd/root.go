package cmd

import (
	"github.com/hashload/boss/cmd/config"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/setup"
	"github.com/hashload/boss/utils"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "boss",
	Short: "Dependency Manager for Delphi",
	Long:  "Dependency Manager for Delphi",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	RootCmd.PersistentFlags().BoolVarP(&env.Global, "global", "g", false, "global environment")
	RootCmd.PersistentFlags().BoolVarP(&msg.DebugEnable, "debug", "d", false, "debug")
	msg.DebugEnable = utils.Contains(os.Args, "-d")

	setup.Initialize()

	config.InitilizeConfig(RootCmd)

	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
