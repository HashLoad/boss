package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/spf13/cobra"
	"io"
	"os/exec"
	"strings"
)

var runScript = &cobra.Command{
	Use:   "run",
	Short: "Run cmd script",
	Long:  `Run cmd script`,
	Run: func(cmd *cobra.Command, args []string) {
		if pkgJson, e := models.LoadPackage(true); e != nil {
			e.Error()
		} else {
			scripts := pkgJson.Scripts.(map[string]interface{})
			if command, ok := scripts[args[0]]; !ok {
				errors.New("script not exists").Error()
			} else {
				runCmd(command.(string) + " " + strings.Join(args[1:], " "))
			}
		}
	},
}

func runCmd(cmdName string) {
	//cmdName = "cmd /c " + cmdName

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

func init() {
	RootCmd.AddCommand(runScript)
}
