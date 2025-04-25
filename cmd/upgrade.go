package cmd

import (
	"github.com/hashload/boss/internal/upgrade"
	"github.com/hashload/boss/pkg/msg"
	"github.com/spf13/cobra"
)

func upgradeCmdRegister(root *cobra.Command) {
	var preRelease bool

	var upgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the client version",
		Example: `  Upgrade boss:
  boss upgrade

  Upgrade boss with pre-release:
  boss upgrade --dev`,
		Run: func(_ *cobra.Command, _ []string) {
			if err := upgrade.BossUpgrade(preRelease); err != nil {
				msg.Fatal("Failed to upgrade boss: %s", err)
			}
		},
	}

	root.AddCommand(upgradeCmd)
	upgradeCmd.Flags().BoolVar(&preRelease, "dev", false, "pre-release")
}
