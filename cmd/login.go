package cmd

import (
	"fmt"
	"log"
	"os/user"
	"path/filepath"
	"syscall"

	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var removeLogin bool

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Register login to repo",
	Long:  `Register login to repo`,
	Run: func(cmd *cobra.Command, args []string) {
		configuration := models.GlobalConfiguration

		if removeLogin {
			delete(configuration.Auth, args[0])
			configuration.SaveConfiguration()
			return
		}

		var auth *models.Auth
		var repo string
		if len(args) > 0 && args[0] != "" {
			repo = args[0]
			auth = configuration.Auth[args[0]]
		} else {
			repo = getParamOrDef("Url to logun (ex: github.com)", "")
			if repo == "" {
				msg.Die("Empty is not valid!!")
			}
			auth = configuration.Auth[repo]
		}

		if auth == nil {
			auth = &models.Auth{}
		}

		auth.UseSsh = getParamBoolean("Use SSH")
		if auth.UseSsh {
			auth.Path = getParamOrDef("Patch of ssh private key("+getSshKeyPath()+")", getSshKeyPath())
		} else {
			auth.SetUser(getParamOrDef("Username", ""))
			auth.SetPass(getPass("Password"))
		}
		configuration.Auth[repo] = auth
		configuration.SaveConfiguration()
	},
}

func getPass(description string) string {
	fmt.Print(description + ": ")

	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
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

func getParamBoolean(msgS string) bool {
	msg.Print(msgS + "(y or n): ")

	return msg.PromptUntilYorN()
}

func init() {
	loginCmd.Flags().BoolVarP(&removeLogin, "rm", "r", false, "remove login")
	RootCmd.AddCommand(loginCmd)
}
