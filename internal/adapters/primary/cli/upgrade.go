// Package cli implements Boss CLI commands.
package cli

import (
	"github.com/hashload/boss/internal/upgrade"
	"github.com/spf13/cobra"
)

// upgradeCmdRegister registers the upgrade command.
func upgradeCmdRegister(root *cobra.Command) {
	var preRelease bool

	var upgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the client version",
		Example: `  Upgrade boss:
  boss upgrade

  Upgrade boss with pre-release:
  boss upgrade --dev`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return upgrade.BossUpgrade(preRelease)
		},
	}

	root.AddCommand(upgradeCmd)
	upgradeCmd.Flags().BoolVar(&preRelease, "dev", false, "pre-release")
}
