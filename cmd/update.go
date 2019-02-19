package cmd

import (
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update dependencies",
	Long:  `update dependencies`,
	Run: func(cmd *cobra.Command, args []string) {
		installCmd.Run(cmd, []string{})
	},
}

func init() {
	RootCmd.AddCommand(updateCmd)
}
