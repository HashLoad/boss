package cmd

import (
	"github.com/hashload/boss/models"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

var selectDelphi = &cobra.Command{
	Use:   "switch delphi",
	Short: "Switch Delphi version",
	Long:  `Switch Delphi version to compile modules`,
	Run: func(cmd *cobra.Command, args []string) {

		config := models.GlobalConfiguration

		prompt := &survey.Select{
			Message: "Choose a Delphi installation:",
			Options: []string{"20", "19", "18"},
			Default: config.DelphiPath,
		}

		survey.AskOne(prompt, &config.DelphiPath, nil)
		config.SaveConfiguration()

	},
}

func init() {
	//RootCmd.AddCommand(selectDelphi)
}
