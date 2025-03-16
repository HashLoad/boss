package cmd

import (
	"github.com/hashload/boss/pkg/installer"
	"github.com/spf13/cobra"
)

var noSaveInstall bool

var installCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install a new dependency",
	Long:    `This command install a new dependency on your project`,
	Aliases: []string{"i", "add"},
	Example: `  Add a new dependency:
  boss install <pkg>

  Add a new version-specific dependency:
  boss install <pkg>@<version>

  Install a dependency without add it from the boss.json file:
  boss install <pkg> --no-save`,
	Run: func(cmd *cobra.Command, args []string) {
		installer.InstallModules(args, true, noSaveInstall)
	},
}

func init() {
	root.AddCommand(installCmd)
	installCmd.Flags().BoolVar(&noSaveInstall, "no-save", false, "prevents saving to dependencies")
}
