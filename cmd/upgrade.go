package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var preRelease bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade the client version",
	Example: `  Upgrade boss:
  boss upgrade

  Upgrade boss with pre-release:
  boss upgrade --dev`,
	Run: func(cmd *cobra.Command, args []string) {
		core.DoBossUpgrade(preRelease)
	},
}

func init() {
	RootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().BoolVar(&preRelease, "dev", false, "pre-release")
}
