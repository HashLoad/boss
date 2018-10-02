package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/hashload/boss/core/gb"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/utils"
	"github.com/spf13/cobra"
	"strings"
)

var dev bool

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a dependency",
	Long:  `Install a dependency`,
	Run: func(cmd *cobra.Command, args []string) {
		loadPackage, e := models.LoadPackage(true)
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
		if loadPackage.IsNew && len(args) == 0 {
			return
		}
		loadPackage.Save()
		core.EnsureDependencies(loadPackage)
		utils.UpdateLibraryPath()
		gb.RunGB()
	},
}

func init() {
	installCmd.Flags().BoolVarP(&dev, "dev", "d", false, "add dev mode")
	RootCmd.AddCommand(installCmd)
}
