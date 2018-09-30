package cmd

import (
	"bytes"
	"errors"
	"github.com/hashload/boss/models"
	"github.com/spf13/cobra"
	"log"
	"os/exec"
	"strings"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run cmd script",
	Long:  `Run cmd script`,
	Run: func(cmd *cobra.Command, args []string) {
		if pkgJson, e := models.LoadPackage(true); e != nil {
			e.Error()
		} else {
			scripts := pkgJson.Scripts.(map[string]interface{})
			if command, ok := scripts[args[0]]; !ok {
				errors.New("Script not exists!").Error()
			} else {
				splited := strings.Split(strings.Trim(command.(string), ""), " ")
				cmd := exec.Command(splited[0], splited[1:]...)
				var out bytes.Buffer
				cmd.Stdout = &out
				err := cmd.Run()
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
}
