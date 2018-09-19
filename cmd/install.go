package cmd

import (
	"github.com/hashload/boss/models"
	"github.com/spf13/cobra"
	"strings"
)

var dev bool

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a dependency",
	Long:  `Install a dependency`,
	Run: func(cmd *cobra.Command, args []string) {
		loadPackage, e := models.LoadPackage("package.json")
		if e != nil {
			e.Error()
		}

		for e := range args {
			dependency := args[e]
			split := strings.Split(dependency, ":")
			if dev {
				loadPackage.AddDevDependency(split[0], split[1])
			} else {
				loadPackage.AddDependency(split[0], split[1])
			}
		}
		loadPackage.Save()
	},
}

func init() {
	installCmd.Flags().BoolVar(&dev, "d", false, "add dev mode")
	installCmd.Flags().BoolVar(&dev, "dev", false, "add dev mode")
	RootCmd.AddCommand(installCmd)
}
