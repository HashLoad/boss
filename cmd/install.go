package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/hashload/boss/core/gb"
	"github.com/hashload/boss/core/paths"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
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
			msg.Die("Fail on open boss.json: %s", e)
		}

		for e := range args {
			dependency := args[e]
			split := strings.Split(dependency, ":")
			var ver string
			if len(split) == 1 {
				ver = "^1"
			} else {
				ver = split[1]
			}
			if dev {

				loadPackage.AddDevDependency(split[0], ver)
			} else {
				loadPackage.AddDependency(split[0], ver)
			}
		}
		if loadPackage.IsNew && len(args) == 0 {
			return
		}
		loadPackage.Save()
		paths.EnsureModulesDir()
		core.EnsureDependencies(loadPackage)
		utils.UpdateLibraryPath()
		gb.RunGB()
	},
}

func init() {
	installCmd.Flags().BoolVarP(&dev, "dev", "d", false, "add dev mode")
	RootCmd.AddCommand(installCmd)
}
