package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils/dcc32"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

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
			if _, err := strconv.Atoi(args[0]); err != nil {
				if _, err = os.Stat(args[0]); os.IsNotExist(err) {
					return errors.New("invalid path")
				}
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

func selectDelphiInteractive() {
	paths := dcc32.GetDcc32DirByCmd()
	if len(paths) == 0 {
		msg.Warn("No Delphi installations found in $PATH")
		msg.Info("You can manually specify a path using: boss config delphi use <path>")
		return
	}

	currentPath := env.GlobalConfiguration().DelphiPath

	options := make([]string, len(paths))
	defaultIndex := 0
	for i, path := range paths {
		if path == currentPath {
			options[i] = fmt.Sprintf("%s (current)", path)
			defaultIndex = i
		} else {
			options[i] = path
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
	config.DelphiPath = paths[selectedIndex]
	config.SaveConfiguration()

	msg.Info("✓ Delphi version updated successfully!")
	msg.Info("  Path: %s", paths[selectedIndex])
}

func listDelphiVersions() {
	paths := dcc32.GetDcc32DirByCmd()
	if len(paths) == 0 {
		msg.Warn("Installations not found in $PATH")
		return
	}

	currentPath := env.GlobalConfiguration().DelphiPath
	msg.Warn("Installations found:")
	for index, path := range paths {
		if path == currentPath {
			msg.Info("  [%d] %s (current)", index, path)
		} else {
			msg.Info("  [%d] %s", index, path)
		}
	}
}

func useDelphiVersion(pathOrIndex string) {
	config := env.GlobalConfiguration()
	if index, err := strconv.Atoi(pathOrIndex); err == nil {
		delphiPaths := dcc32.GetDcc32DirByCmd()
		if index < 0 || index >= len(delphiPaths) {
			msg.Die("Invalid index: %d. Use 'boss config delphi list' to see available options", index)
		}
		config.DelphiPath = delphiPaths[index]
	} else {
		config.DelphiPath = pathOrIndex
	}

	config.SaveConfiguration()
	msg.Info("Successful!")
}
