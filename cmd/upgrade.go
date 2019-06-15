package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var preRelease bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade a cli",
	Long:  `upgrade a cli`,
	Run: func(cmd *cobra.Command, args []string) {
		core.DoBossUpgrade(preRelease)
	},
}

func init() {
	RootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().BoolVar(&preRelease, "dev", false, "Pre-release")
}
