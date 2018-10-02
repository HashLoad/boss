package cmd

import (
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/spf13/cobra"
	"log"
	"os/user"
	"path/filepath"
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
			auth.SetPass(getParamOrDef("Password", ""))
		}
		configuration.Auth[repo] = auth
		configuration.SaveConfiguration()
	},
}

func getSshKeyPath() string {
	usr, e := user.Current()
	if e != nil {
		log.Fatal(e)
	}
	return filepath.Join(usr.HomeDir, ".ssh", "id_rsa")
}

func getParamBoolean(msgS string) bool {
	msg.Print(msgS + "(y for yes): ")

	return msg.PromptUntilYorN()
}

func init() {
	loginCmd.Flags().BoolVarP(&removeLogin, "rm", "r", false, "remove login")
	RootCmd.AddCommand(loginCmd)
}
