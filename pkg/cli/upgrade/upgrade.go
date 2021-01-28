package upgrade

import (
	"github.com/hashload/boss/pkg/util"
	"github.com/spf13/cobra"
)

// NewCmdUpgrade add the command line upgrage
func NewCmdUpgrade() *cobra.Command {
	var preRelease bool
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the client version",
		Example: `  Upgrade boss:
  boss upgrade

  Upgrade boss with pre-release:
  boss upgrade --dev`,
		Run: func(cmd *cobra.Command, args []string) {
			err := upgradeBoss(preRelease)
			util.CheckErr(err)
		},
	}
	cmd.Flags().BoolVar(&preRelease, "dev", false, "pre-release")
	return cmd
}

func upgradeBoss(preRelease bool) error {
	return nil
}