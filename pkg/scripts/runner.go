package scripts

import (
	"bufio"
	"errors"
	"io"
	"os/exec"

	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
)

func RunCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmdReader, err := cmd.StdoutPipe()
	cmdErr, _ := cmd.StderrPipe()
	if err != nil {
		msg.Err("Error creating StdoutPipe for Cmd", err)
		return
	}
	merged := io.MultiReader(cmdReader, cmdErr)
	scanner := bufio.NewScanner(merged)
	go func() {
		for scanner.Scan() {
			text := scanner.Text()
			msg.Info("%s\n", text)
		}
	}()

	err = cmd.Start()
	if err != nil {
		msg.Err("Error starting Cmd", err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		msg.Err("Error waiting for Cmd", err)
		return
	}
}

func Run(args []string) {
	if packageData, err := models.LoadPackage(true); err != nil {
		msg.Err(err.Error())
	} else {
		if packageData.Scripts == nil {
			msg.Die(errors.New("script not exists").Error())
		}

		if command, ok := packageData.Scripts[args[0]]; !ok {
			msg.Err(errors.New("script not exists").Error())
		} else {
			RunCmd(command, args[1:]...)
		}
	}
}
