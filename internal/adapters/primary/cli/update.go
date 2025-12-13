package cli

import (
	"fmt"
	"os"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/installer"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func updateCmdRegister(root *cobra.Command) {
	var selectMode bool

	var updateCmd = &cobra.Command{
		Use:     "update",
		Short:   "Update dependencies",
		Long:    `This command update installed dependencies`,
		Aliases: []string{"up"},
		Example: `  Update all dependencies:
  boss update

  Select specific dependencies to update:
  boss update --select`,
		Run: func(_ *cobra.Command, args []string) {
			if selectMode {
				updateWithSelect()
			} else {
				installer.InstallModules(installer.InstallOptions{
					Args:          args,
					LockedVersion: false,
					NoSave:        false,
				})
			}
		},
	}

	updateCmd.Flags().BoolVarP(&selectMode, "select", "s", false, "select dependencies to update")
	root.AddCommand(updateCmd)
}

func updateWithSelect() {
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

		if installed.Version == "" {
			options[i] = fmt.Sprintf("%s (not installed)", dep.Name())
		} else if dep.GetVersion() != installed.Version {
			options[i] = fmt.Sprintf("%s (%s → %s)", dep.Name(), installed.Version, dep.GetVersion())
		} else {
			options[i] = fmt.Sprintf("%s (up to date)", dep.Name())
		}
	}

	selectedOptions, err := pterm.DefaultInteractiveMultiselect.
		WithOptions(options).
		WithDefaultText("Select dependencies to update (Space to select, Enter to confirm):").
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

	msg.Info("Updating %d dependencies...\n", len(selectedDeps))
	installer.InstallModules(installer.InstallOptions{
		Args:          selectedDeps,
		LockedVersion: true,
		NoSave:        false,
		ForceUpdate:   selectedDeps,
	})
}
