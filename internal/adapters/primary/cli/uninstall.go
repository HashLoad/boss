package cli

import (
	"os"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/installer"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// uninstallCmdRegister registers the uninstall command
func uninstallCmdRegister(root *cobra.Command) {
	var noSaveUninstall bool
	var selectMode bool

	var uninstallCmd = &cobra.Command{
		Use:     "uninstall",
		Short:   "Uninstall a dependency",
		Long:    "This uninstalls a package, completely removing everything boss installed on its behalf",
		Aliases: []string{"remove", "rm", "r", "un", "unlink"},
		Example: `  Uninstall a package:
  boss uninstall <pkg>

  Uninstall a package without removing it from the boss.json file:
  boss uninstall <pkg> --no-save
  
  Select multiple packages to uninstall:
  boss uninstall --select`,
		Run: func(_ *cobra.Command, args []string) {
			if selectMode {
				uninstallWithSelect(noSaveUninstall)
			} else {
				installer.UninstallModules(args, noSaveUninstall)
			}
		},
	}

	root.AddCommand(uninstallCmd)
	uninstallCmd.Flags().BoolVar(
		&noSaveUninstall,
		"no-save",
		false,
		"package will not be removed from your boss.json file",
	)
	uninstallCmd.Flags().BoolVarP(&selectMode, "select", "s", false, "select dependencies to uninstall")
}

// uninstallWithSelect uninstalls the selected dependencies
func uninstallWithSelect(noSave bool) {
	pkg, err := domain.LoadPackage(false)
	if err != nil {
		if os.IsNotExist(err) {
			msg.Die("boss.json not exists in " + env.GetCurrentDir())
		} else {
			msg.Die("Fail on open dependencies file: %s", err)
		}
	}

	deps := pkg.GetParsedDependencies()
	if len(deps) == 0 {
		msg.Info("No dependencies found in boss.json")
		return
	}

	options := make([]string, len(deps))
	depNames := make([]string, len(deps))

	for i, dep := range deps {
		depNames[i] = dep.Repository
		installed := pkg.Lock.GetInstalled(dep)

		if installed.Version != "" {
			options[i] = dep.Name() + " (installed)"
		} else {
			options[i] = dep.Name() + " (not installed)"
		}
	}

	selectedOptions, err := pterm.DefaultInteractiveMultiselect.
		WithOptions(options).
		WithDefaultText("Select dependencies to remove (Space to select, Enter to confirm):").
		Show()

	if err != nil {
		msg.Die("Error selecting dependencies: %s", err)
	}

	if len(selectedOptions) == 0 {
		msg.Info("No dependencies selected")
		return
	}

	selectedDeps := make([]string, 0, len(selectedOptions))
	for _, selected := range selectedOptions {
		for i, opt := range options {
			if opt == selected {
				selectedDeps = append(selectedDeps, depNames[i])
				break
			}
		}
	}

	msg.Info("Uninstalling %d dependencies...\n", len(selectedDeps))
	installer.UninstallModules(selectedDeps, noSave)
}
