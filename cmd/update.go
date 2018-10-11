package cmd

import (
	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "update",
	Short: "update a cli",
	Long:  `update a cli`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	//RootCmd.AddCommand(publishCmd)
}
