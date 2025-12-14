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

	setup.Initialize()

	config.RegisterConfigCommand(root)
	initCmdRegister(root)
	installCmdRegister(root)
	loginCmdRegister(root)
	runCmdRegister(root)
	uninstallCmdRegister(root)
	updateCmdRegister(root)
	upgradeCmdRegister(root)
	dependenciesCmdRegister(root)
	versionCmdRegister(root)

	if err := gc.RunGC(false); err != nil {
		return err
	}

	config.RegisterCmd(root)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}

	return nil
}
