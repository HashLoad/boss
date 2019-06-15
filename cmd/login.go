package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var removeLogin bool

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Register login to repo",
	Long:  `Register login to repo`,
	Run: func(cmd *cobra.Command, args []string) {
		core.Login()
	},
}

func init() {
	loginCmd.Flags().BoolVarP(&removeLogin, "rm", "r", false, "remove login")
	RootCmd.AddCommand(loginCmd)
}
