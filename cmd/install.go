package cmd

import (
	"github.com/hashload/boss/core/compiler"
	"strings"

	"github.com/hashload/boss/core"
	"github.com/hashload/boss/core/gc"
	"github.com/hashload/boss/core/paths"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"github.com/spf13/cobra"
)

var dev bool

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a dependency",
	Long:  `Install a dependency`,
	Run: func(cmd *cobra.Command, args []string) {
		loadPackage, e := models.LoadPackage(false)
		if e != nil {
			msg.Die("Fail on open dependencies file: %s", e)
		}

		for e := range args {
			dependency := args[e]
			split := strings.Split(dependency, ":")
			var ver string
			if len(split) == 1 {
				ver = "x"
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
		paths.EnsureModulesDir()
		msg.Info("Installing modules in project patch")
		core.EnsureDependencies(loadPackage)
		utils.UpdateLibraryPath()
		msg.Info("Compiling units")
		compiler.BuildDucs()
		loadPackage.Save()
		gc.RunGC()
	},
}

func init() {
	installCmd.Flags().BoolVarP(&dev, "dev", "d", false, "add dev mode")
	RootCmd.AddCommand(installCmd)
}
