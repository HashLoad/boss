package cmd

import (
	"github.com/hashload/boss/pkg/installer"
	"github.com/spf13/cobra"
)

func uninstallCmdRegister(root *cobra.Command) {
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
		Run: func(_ *cobra.Command, args []string) {
			installer.UninstallModules(args, noSaveUninstall)
		},
	}

	root.AddCommand(uninstallCmd)
	uninstallCmd.Flags().BoolVar(
		&noSaveUninstall,
		"no-save",
		false,
		"package will not be removed from your boss.json file",
	)
}
