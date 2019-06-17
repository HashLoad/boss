package scripts

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"io"
	"os/exec"
	"strings"
)

func RunCmd(cmdName string) {
	fields := strings.Fields(cmdName)

	cmd := exec.Command(fields[0], fields[1:]...)
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
			fmt.Printf("%s\n", text)
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
	if pkgJson, e := models.LoadPackage(true); e != nil {
		e.Error()
	} else {
		scripts := pkgJson.Scripts.(map[string]interface{})
		if command, ok := scripts[args[0]]; !ok {
			msg.Err(errors.New("script not exists").Error())
		} else {
			RunCmd(command.(string) + " " + strings.Join(args[1:], " "))
		}
	}

}
