package cmd

import (
	"github.com/hashload/boss/env"
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

	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
