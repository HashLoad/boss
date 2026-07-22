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

const (
	appName        = "boss"
	appDescription = "Dependency Manager for Delphi"
)

// Command names, shared between each command registration and the grouping
// pass in applyCommandGroups, which matches commands by name.
const (
	cmdNameNew        = "new"
	cmdNameRun        = "run"
	cmdNameLogin      = "login"
	cmdNameWorkspace  = "workspace"
	cmdNameContribute = "contribute"
	cmdNameCRA        = "cra"
	cmdNameVersion    = "version"
)

// Identifiers of the help groups printed by 'boss --help'.
const (
	groupIDLegacy    = "legacy"
	groupIDProject   = "new"
	groupIDPubPascal = "pubpascal"
	groupIDCRA       = "cra"
)

// flagNameVersion is the long form of the --version flag.
const flagNameVersion = "version"

// Execute executes the root command.
func Execute() error {
	var versionPrint bool
	var global bool
	var debug bool

	var root = &cobra.Command{
		Use:   appName,
		Short: appDescription,
		Long:  appDescription,
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
	root.Flags().BoolVarP(&versionPrint, flagNameVersion, "v", false, "show cli version")

	isHelpOrVersion := isHelpOrVersionInvocation(os.Args)

	if isHelpOrVersion {
		setup.InitializeMinimal()
	} else {
		setup.Initialize()
	}

	registerCommands(root)
	applyCommandGroups(root)

	if !isHelpOrVersion {
		if err := gc.RunGC(false); err != nil {
			return err
		}
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}

	return nil
}

// isHelpOrVersionInvocation reports whether boss was called only to print help
// or the version, in which case the full environment setup can be skipped.
func isHelpOrVersionInvocation(args []string) bool {
	if len(args) <= 1 {
		return true
	}

	for _, arg := range args[1:] {
		switch arg {
		case "help", "-h", "--help", cmdNameVersion, "-v", "--version":
			return true
		}
	}

	return false
}

// registerCommands wires every boss command into the root command.
func registerCommands(root *cobra.Command) {
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

	// Registered before the grouping pass in applyCommandGroups: any command
	// added afterwards keeps an empty GroupID and cobra prints it in a stray
	// "Additional Commands" block instead of one of the groups.
	config.RegisterCmd(root)
}

// applyCommandGroups declares the help groups and assigns every registered
// command to one of them.
func applyCommandGroups(root *cobra.Command) {
	root.AddGroup(
		&cobra.Group{ID: groupIDLegacy, Title: "Available Commands:"},
		&cobra.Group{ID: groupIDProject, Title: "Project & Packaging:"},
		&cobra.Group{ID: groupIDPubPascal, Title: "PubPascal Portal:"},
		&cobra.Group{ID: groupIDCRA, Title: "Cyber Resilience Act (CRA) & SBOM:"},
	)

	for _, cmd := range root.Commands() {
		switch cmd.Name() {
		case cmdNameNew, projectTypePkg, cmdNameRun:
			cmd.GroupID = groupIDProject
		case cmdNameLogin, cmdNameWorkspace, cmdNameContribute:
			cmd.GroupID = groupIDPubPascal
		case cmdNameCRA, sbomBaseName:
			cmd.GroupID = groupIDCRA
		default:
			cmd.GroupID = groupIDLegacy
		}
	}
}
