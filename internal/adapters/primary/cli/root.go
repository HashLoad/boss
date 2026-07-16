// Package cli provides the command-line interface for Boss package manager.
// It implements commands for dependency management, build operations, and configuration.
package cli

import (
	"os"

	"github.com/hashload/boss/internal/adapters/primary/cli/config"
	"github.com/hashload/boss/internal/core/services/gc"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/setup"

	"github.com/spf13/cobra"
)

// Execute executes the root command.
func Execute() error {
	var versionPrint bool
	var global bool
	var debug bool

	var root = &cobra.Command{
		Use:   "boss",
		Short: "Dependency Manager for Delphi",
		Long:  "Dependency Manager for Delphi",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if debug {
				msg.LogLevel(msg.DEBUG)
				msg.Debug("Debug mode enabled")
			}

			env.SetGlobal(global)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if versionPrint {
				printVersion()
			} else {
				return cmd.Help()
			}
			return nil
		},
	}

	root.PersistentFlags().BoolVarP(&global, "global", "g", false, "global environment")
	root.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug")
	root.Flags().BoolVarP(&versionPrint, "version", "v", false, "show cli version")

	isHelpOrVersion := false
	if len(os.Args) <= 1 {
		isHelpOrVersion = true
	} else {
		for _, arg := range os.Args[1:] {
			if arg == "help" || arg == "-h" || arg == "--help" || arg == "version" || arg == "-v" || arg == "--version" {
				isHelpOrVersion = true
				break
			}
		}
	}

	if isHelpOrVersion {
		setup.InitializeMinimal()
	} else {
		setup.Initialize()
	}

	config.RegisterConfigCommand(root)
	initCmdRegister(root)
	newCmdRegister(root)
	installCmdRegister(root)
	loginCmdRegister(root)
	runCmdRegister(root)
	uninstallCmdRegister(root)
	updateCmdRegister(root)
	upgradeCmdRegister(root)
	dependenciesCmdRegister(root)
	versionCmdRegister(root)
	pubpascalCmdRegister(root)
	craCmdRegister(root)
	contributeCmdRegister(root)

	legacyGroup := &cobra.Group{
		ID:    "legacy",
		Title: "Available Commands:",
	}
	newGroup := &cobra.Group{
		ID:    "new",
		Title: "Available Commands (new):",
	}
	pubpascalGroup := &cobra.Group{
		ID:    "pubpascal",
		Title: "Available Commands (pubpascal):",
	}
	craGroup := &cobra.Group{
		ID:    "cra",
		Title: "Cyber Resilience Act (CRA) & SBOM:",
	}

	root.AddGroup(legacyGroup, newGroup, pubpascalGroup, craGroup)

	for _, cmd := range root.Commands() {
		switch cmd.Name() {
		case "new", "pkg", "run":
			cmd.GroupID = "new"
		case "login", "workspace", "contribute":
			cmd.GroupID = "pubpascal"
		case "cra", "sbom", "scan", "publish-sbom":
			cmd.GroupID = "cra"
		default:
			cmd.GroupID = "legacy"
		}
	}

	if !isHelpOrVersion {
		if err := gc.RunGC(false); err != nil {
			return err
		}
	}

	config.RegisterCmd(root)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}

	return nil
}
