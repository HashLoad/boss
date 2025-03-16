package cmd

import (
	"fmt"
	"log"
	"os/user"
	"path/filepath"
	"syscall"

	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var removeLogin bool
var useSsh bool
var privateKey string
var userName string
var password string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Add a registry user account",
	Example: `  Adding a new user account:
  boss login <repo>`,
	Aliases: []string{"adduser", "add-user"},
	Run: func(cmd *cobra.Command, args []string) {
		login(removeLogin, useSsh, privateKey, userName, password, args)
	},
}

func init() {
	// TODO add example to remove login or add a new command to logout (equals branch refact-steroids)
	loginCmd.Flags().BoolVarP(&removeLogin, "rm", "r", false, "remove login")
	loginCmd.Flags().BoolVarP(&useSsh, "ssh", "s", false, "Use SSH")
	loginCmd.Flags().StringVarP(&privateKey, "key", "k", "", "Path of ssh private key")
	loginCmd.Flags().StringVarP(&userName, "username", "u", "", "Username")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Password or PassPhrase(with SSH)")
	root.AddCommand(loginCmd)
}

func login(removeLogin bool, useSSH bool, privateKey string, userName string, password string, args []string) {
	configuration := env.GlobalConfiguration

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

func setAuthWithParams(auth *env.Auth, useSsh bool, privateKey, userName, password string) {
	auth.UseSSH = useSsh
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
		auth.Path = getParamOrDef("Path of ssh private key("+getSshKeyPath()+")", getSshKeyPath())
		auth.SetPassPhrase(getPass("PassPhrase"))
	} else {
		auth.SetUser(getParamOrDef("Username", ""))
		auth.SetPass(getPass("Password"))
	}
}

func getPass(description string) string {
	fmt.Print(description + ": ")

	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		msg.Die("Error on get pass: %s", err)
	}
	return string(bytePassword)
}

func getSshKeyPath() string {
	usr, e := user.Current()
	if e != nil {
		log.Fatal(e)
	}
	return filepath.Join(usr.HomeDir, ".ssh", "id_rsa")
}

func getParamBoolean(msg string) bool {
	result, _ := pterm.DefaultInteractiveConfirm.Show(msg)
	return result
}
