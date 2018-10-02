package cmd

import (
	"github.com/hashload/boss/msg"

	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "publish a dependency",
	Long:  `publish a dependency`,
	Run: func(cmd *cobra.Command, args []string) {
		msg.Err("TODO")
	},
}

func init() {
	RootCmd.AddCommand(publishCmd)
}
