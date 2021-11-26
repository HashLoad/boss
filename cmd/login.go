package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var removeLogin bool

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Add a registry user account",
	Example: `  Adding a new user account:
  boss login <repo>`,
	Aliases: []string{"adduser", "add-user"},
	Run: func(cmd *cobra.Command, args []string) {
		core.Login(removeLogin, args)
	},
}

func init() {
	// TODO add example to remove login or add a new command to logout (equals branch refact-steroids)
	loginCmd.Flags().BoolVarP(&removeLogin, "rm", "r", false, "remove login")
	RootCmd.AddCommand(loginCmd)
}
