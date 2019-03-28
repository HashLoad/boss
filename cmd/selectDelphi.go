package cmd

import (
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
	"os/exec"
	"path/filepath"
	"strings"
)

var selectDelphi = &cobra.Command{
	Use:     "switch",
	Short:   "Switch Delphi version",
	Long:    `Switch Delphi version to compile modules`,
	Aliases: []string{"sd"},
	Run: func(cmd *cobra.Command, args []string) {

		config := models.GlobalConfiguration

		prompt := &survey.Select{
			Message: "Choose a Delphi installation:",
			Options: getDcc32Dir(),
			Default: config.DelphiPath,
		}

		survey.AskOne(prompt, &config.DelphiPath, nil)
		config.SaveConfiguration()

	},
}

func getDcc32Dir() []string {
	command := exec.Command("where", "dcc32")
	output, err := command.Output()
	if err != nil {
		msg.Warn("dcc32 not found")
	}
	outputStr := strings.ReplaceAll(string(output), "\t", "")
	outputStr = strings.ReplaceAll(outputStr, "\r", "")
	if strings.HasSuffix(outputStr, "\n") {
		outputStr = outputStr[0 : len(outputStr)-1]
	}
	instalations := strings.Split(outputStr, "\n")
	for key, value := range instalations {
		instalations[key] = filepath.Dir(value)
	}
	return instalations
}

func init() {
	RootCmd.AddCommand(selectDelphi)
}
