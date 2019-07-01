package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var preRelease bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade a cli",
	Long:  `This command upgrade the client version`,
	Run: func(cmd *cobra.Command, args []string) {
		core.DoBossUpgrade(preRelease)
	},
}

func init() {
	RootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().BoolVar(&preRelease, "dev", false, "Pre-release")
}
