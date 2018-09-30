package cmd

import (
	"fmt"

	"github.com/hashload/boss/models"
	"github.com/spf13/cobra"
)

func publish(p *models.Package) {
	fmt.Println(p.Homepage)
}

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "publish a dependency",
	Long:  `publish a dependency`,
	Run: func(cmd *cobra.Command, args []string) {
		loadPackage, e := models.LoadPackage( false)
		if e != nil {
			e.Error()
		}

		publish(loadPackage)

		loadPackage.Save()
	},
}

func init() {
	RootCmd.AddCommand(publishCmd)
}
