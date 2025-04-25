package scripts

import (
	"os"
	"os/exec"
	"strings"

	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
)

func executeCommand(name string, args ...string) {
	cmdParts := strings.Fields(name)
	name = cmdParts[0]
	args = append(cmdParts[1:], args...)

	cmd := exec.Command(name, args...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		msg.Err("failed to execute command: %s", err.Error())
		msg.Err("command: %s", name)
		msg.Err("args: %s", strings.Join(args, " "))
		return
	}
}

func Run(args []string) {
	if packageData, err := models.LoadPackage(true); err != nil {
		msg.Err(err.Error())
	} else {
		if packageData.Scripts == nil {
			msg.Fatal("script not exists")
		}

		if command, ok := packageData.Scripts[args[0]]; !ok {
			msg.Err("script not exists")
		} else {
			executeCommand(command, args[1:]...)
		}
	}
}
