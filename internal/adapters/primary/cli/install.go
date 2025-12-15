// Package cli provides command-line interface implementation for Boss.
package cli

import (
	"github.com/hashload/boss/internal/core/services/installer"
	"github.com/spf13/cobra"
)

// installCmdRegister registers the install command.
func installCmdRegister(root *cobra.Command) {
	var noSaveInstall bool
	var compilerVersion string
	var platform string
	var strict bool

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
  boss install <pkg> --no-save

  Install using a specific compiler version:
  boss install --compiler=35.0

  Install using a specific platform:
  boss install --platform=Win64`,
		Run: func(_ *cobra.Command, args []string) {
			installer.InstallModules(installer.InstallOptions{
				Args:          args,
				LockedVersion: true,
				NoSave:        noSaveInstall,
				Compiler:      compilerVersion,
				Platform:      platform,
				Strict:        strict,
			})
		},
	}

	root.AddCommand(installCmd)
	installCmd.Flags().BoolVar(&noSaveInstall, "no-save", false, "prevents saving to dependencies")
	installCmd.Flags().StringVar(&compilerVersion, "compiler", "", "compiler version to use")
	installCmd.Flags().StringVar(&platform, "platform", "", "platform to use (e.g., Win32, Win64)")
	installCmd.Flags().BoolVar(&strict, "strict", false, "strict mode for compiler selection")
}
