package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	registryadapter "github.com/hashload/boss/internal/adapters/secondary/registry"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// delphiCmd registers the delphi command
func delphiCmd(root *cobra.Command) {
	delphiCmd := &cobra.Command{
		Use:   "delphi",
		Short: "Configure Delphi version",
		Long:  `Configure Delphi version to compile modules`,
		Run: func(cmd *cobra.Command, _ []string) {
			selectDelphiInteractive()
		},
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List Delphi versions",
		Long:  `List Delphi versions to compile modules`,
		Run: func(_ *cobra.Command, _ []string) {
			listDelphiVersions()
		},
	}

	use := &cobra.Command{
		Use:   "use [path]",
		Short: "Use Delphi version",
		Long:  `Use Delphi version to compile modules`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}
			return nil
		},
		Run: func(_ *cobra.Command, args []string) {
			useDelphiVersion(args[0])
		},
	}

	root.AddCommand(delphiCmd)

	delphiCmd.AddCommand(list)
	delphiCmd.AddCommand(use)
}

// selectDelphiInteractive selects the delphi version interactively
func selectDelphiInteractive() {
	installations := registryadapter.GetDetectedDelphis()
	if len(installations) == 0 {
		msg.Warn("No Delphi installations found in registry")
		msg.Info("You can manually specify a path using: boss config delphi use <path>")
		return
	}

	currentPath := env.GlobalConfiguration().DelphiPath

	options := make([]string, len(installations))
	defaultIndex := 0
	for i, inst := range installations {
		instDir := filepath.Dir(inst.Path)
		label := fmt.Sprintf("%s (%s)", inst.Version, inst.Arch)
		if strings.EqualFold(instDir, currentPath) {
			options[i] = fmt.Sprintf("%s (current)", label)
			defaultIndex = i
		} else {
			options[i] = label
		}
	}

	msg.Info("Current Delphi path: %s\n", currentPath)

	selectedOption, err := pterm.DefaultInteractiveSelect.
		WithOptions(options).
		WithDefaultText("Select Delphi version to use:").
		WithDefaultOption(options[defaultIndex]).
		Show()

	if err != nil {
		msg.Err("Error selecting Delphi version: %s", err)
		return
	}

	selectedIndex := -1
	for i, opt := range options {
		if opt == selectedOption {
			selectedIndex = i
			break
		}
	}

	if selectedIndex == -1 {
		msg.Err("Invalid selection")
		return
	}

	config := env.GlobalConfiguration()
	config.DelphiPath = filepath.Dir(installations[selectedIndex].Path)
	config.SaveConfiguration()

	msg.Info("✓ Delphi version updated successfully!")
	msg.Info("  Path: %s", config.DelphiPath)
}

// listDelphiVersions lists the delphi versions
func listDelphiVersions() {
	installations := registryadapter.GetDetectedDelphis()
	if len(installations) == 0 {
		msg.Warn("Installations not found in registry")
		return
	}

	currentPath := env.GlobalConfiguration().DelphiPath
	msg.Warn("Installations found:")
	for index, inst := range installations {
		instDir := filepath.Dir(inst.Path)
		if strings.EqualFold(instDir, currentPath) {
			msg.Info("  [%d] %s (%s) (current)", index, inst.Version, inst.Arch)
		} else {
			msg.Info("  [%d] %s (%s)", index, inst.Version, inst.Arch)
		}
	}
}

// useDelphiVersion uses the delphi version
func useDelphiVersion(pathOrIndex string) {
	config := env.GlobalConfiguration()
	installations := registryadapter.GetDetectedDelphis()

	if index, err := strconv.Atoi(pathOrIndex); err == nil {
		if index >= 0 && index < len(installations) {
			config.DelphiPath = filepath.Dir(installations[index].Path)
		} else {
			found := false
			for _, inst := range installations {
				if inst.Version == pathOrIndex {
					config.DelphiPath = filepath.Dir(inst.Path)
					found = true
					break
				}

				versionWithArch := fmt.Sprintf("%s-%s", inst.Version, inst.Arch)
				if strings.EqualFold(versionWithArch, pathOrIndex) {
					config.DelphiPath = filepath.Dir(inst.Path)
					found = true
					break
				}
			}
			if !found {
				msg.Die("Invalid index or version: %s. Use 'boss config delphi list' to see available options", pathOrIndex)
			}
		}
	} else {
		found := false
		for _, inst := range installations {

			if inst.Version == pathOrIndex {
				config.DelphiPath = filepath.Dir(inst.Path)
				found = true
				break
			}

			versionWithArch := fmt.Sprintf("%s-%s", inst.Version, inst.Arch)
			if strings.EqualFold(versionWithArch, pathOrIndex) {
				config.DelphiPath = filepath.Dir(inst.Path)
				found = true
				break
			}
		}
		if !found {
			if _, err := os.Stat(pathOrIndex); err == nil {
				config.DelphiPath = pathOrIndex
			} else {
				msg.Die("Invalid version or path: %s", pathOrIndex)
			}
		}
	}

	config.SaveConfiguration()
	msg.Info("Successful!")
}
