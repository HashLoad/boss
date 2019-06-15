package core

import (
	"fmt"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/msg"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os/user"
	"path/filepath"
	"syscall"
)

func Login(removeLogin bool, args []string) {
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

	auth.UseSsh = getParamBoolean("Use SSH")
	if auth.UseSsh {
		auth.Path = getParamOrDef("Path of ssh private key("+getSshKeyPath()+")", getSshKeyPath())
	} else {
		auth.SetUser(getParamOrDef("Username", ""))
		auth.SetPass(getPass("Password"))
	}
	configuration.Auth[repo] = auth
	configuration.SaveConfiguration()

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
