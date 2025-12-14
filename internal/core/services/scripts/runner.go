// Package scripts provides functionality for running custom scripts defined in boss.json.
// It executes shell commands and captures their output for display.
package scripts

import (
	"bufio"
	"errors"
	"io"
	"os/exec"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/msg"
)

// RunCmd executes a command with the given arguments.
func RunCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmdReader, err := cmd.StdoutPipe()
	cmdErr, _ := cmd.StderrPipe()
	if err != nil {
		msg.Err("❌ Error creating StdoutPipe for Cmd", err)
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
		msg.Err("❌ Error starting Cmd", err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		msg.Err("❌ Error waiting for Cmd", err)
		return
	}
}

// Run executes a script defined in the package.
func Run(args []string) {
	if packageData, err := domain.LoadPackage(true); err != nil {
		msg.Err("❌ %s", err.Error())
	} else {
		if packageData.Scripts == nil {
			msg.Die(errors.New("script not exists").Error())
		}

		if command, ok := packageData.Scripts[args[0]]; !ok {
			msg.Err("❌ %s", errors.New("script not exists").Error())
		} else {
			RunCmd(command, args[1:]...)
		}
	}
}
