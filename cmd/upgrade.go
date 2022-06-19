package cmd

import (
	"github.com/hashload/boss/internal/upgrade"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return upgrade.BossUpgrade(preRelease)
	},
}

func init() {
	RootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().BoolVar(&preRelease, "dev", false, "pre-release")
}
