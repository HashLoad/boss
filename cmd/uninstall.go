package cmd

import (
	"github.com/hashload/boss/core"
	"github.com/spf13/cobra"
)

var noSaveUninstall bool

var uninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Short:   "Uninstall a dependency",
	Long:    "This uninstalls a package, completely removing everything boss installed on its behalf",
	Aliases: []string{"remove", "rm", "r", "un", "unlink"},
	Example: `  Uninstall a package:
  boss uninstall <pkg>

  Uninstall a package without removing it from the boss.json file:
  boss uninstall <pkg> --no-save`,
	Run: func(cmd *cobra.Command, args []string) {
		core.UninstallModules(args, noSaveUninstall)
	},
}

func init() {
	RootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().BoolVar(&noSaveUninstall, "no-save", false, "package will not be removed from your boss.json file")
}
