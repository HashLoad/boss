package cmd

import (
	"os/user"
	"path/filepath"

	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func loginCmdRegister(root *cobra.Command) {
	var removeLogin bool
	var useSSH bool
	var privateKey string
	var userName string
	var password string

	var loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Add a registry user account",
		Example: `  Adding a new user account:
  boss login <repo>`,
		Aliases: []string{"adduser", "add-user"},
		Run: func(_ *cobra.Command, args []string) {
			login(removeLogin, useSSH, privateKey, userName, password, args)
		},
	}

	var logoutCmd = &cobra.Command{
		Use:   "logout",
		Short: "Remove a registry user account",
		Example: `  Remove a new user account:
  boss login <repo>`,
		Run: func(_ *cobra.Command, args []string) {
			login(removeLogin, false, "", "", "", args)
		},
	}

	loginCmd.Flags().BoolVarP(&removeLogin, "rm", "r", false, "remove login")
	loginCmd.Flags().BoolVarP(&useSSH, "ssh", "s", false, "Use SSH")
	loginCmd.Flags().StringVarP(&privateKey, "key", "k", "", "Path of ssh private key")
	loginCmd.Flags().StringVarP(&userName, "username", "u", "", "Username")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Password or PassPhrase(with SSH)")
	root.AddCommand(loginCmd)

	root.AddCommand(logoutCmd)
}

func login(removeLogin bool, useSSH bool, privateKey string, userName string, password string, args []string) {
	configuration := env.GlobalConfiguration()

	if removeLogin {
		delete(configuration.Auth, args[0])
		configuration.SaveConfiguration()
		return
	}

	var auth *env.Auth
	var repo string
	if len(args) > 0 && args[0] != "" {
		repo = args[0]
		auth = configuration.Auth[args[0]]
	} else {
		repo = getParamOrDef("Url to login (ex: github.com)", "")
		if repo == "" {
			msg.Die("Empty is not valid!!")
		}
		auth = configuration.Auth[repo]
	}

	if auth == nil {
		auth = &env.Auth{}
	}

	if (userName != "") || (privateKey != "") {
		setAuthWithParams(auth, useSSH, privateKey, userName, password)
	} else {
		setAuthInteractively(auth)
	}

	configuration.Auth[repo] = auth
	configuration.SaveConfiguration()
}

func setAuthWithParams(auth *env.Auth, useSSH bool, privateKey, userName, password string) {
	auth.UseSSH = useSSH
	if auth.UseSSH || (privateKey != "") {
		auth.UseSSH = true
		auth.Path = privateKey
		auth.SetPassPhrase(password)
	} else {
		auth.SetUser(userName)
		auth.SetPass(password)
	}
}

func setAuthInteractively(auth *env.Auth) {
	auth.UseSSH = getParamBoolean("Use SSH")

	if auth.UseSSH {
		auth.Path = getParamOrDef("Path of ssh private key("+getSSHKeyPath()+")", getSSHKeyPath())
		auth.SetPassPhrase(getPass("PassPhrase"))
	} else {
		auth.SetUser(getParamOrDef("Username", ""))
		auth.SetPass(getPass("Password"))
	}
}

func getPass(description string) string {
	pass, err := pterm.DefaultInteractiveTextInput.WithMask("•").Show(description)
	if err != nil {
		msg.Die("Error on get pass: %s", err)
	}
	return pass
}

func getSSHKeyPath() string {
	usr, err := user.Current()
	if err != nil {
		msg.Die(err.Error())
	}
	return filepath.Join(usr.HomeDir, ".ssh", "id_rsa")
}

func getParamBoolean(msg string) bool {
	result, _ := pterm.DefaultInteractiveConfirm.Show(msg)
	return result
}
