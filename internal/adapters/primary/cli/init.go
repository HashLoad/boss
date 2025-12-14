// Package cli provides command-line interface implementation for Boss.
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"

	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/pkg/pkgmanager"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var reFolderName = regexp.MustCompile(`^.+` + regexp.QuoteMeta(string(filepath.Separator)) + `([^\\]+)$`)

// initCmdRegister registers the init command
func initCmdRegister(root *cobra.Command) {
	var quiet bool

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a new project",
		Long:  "Initialize a new project and creates a boss.json file",
		Example: `  Initialize a new project:
  boss init

  Initialize a new project without having it ask any questions:
  boss init --quiet`,
		Run: func(_ *cobra.Command, _ []string) {
			doInitialization(quiet)
		},
	}

	initCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "without asking questions")

	root.AddCommand(initCmd)
}

// doInitialization initializes the project
func doInitialization(quiet bool) {
	if !quiet {
		printHead()
	}

	packageData, err := pkgmanager.LoadPackage()
	if err != nil && !os.IsNotExist(err) {
		msg.Die("Fail on open dependencies file: %s", err)
	}

	allString := reFolderName.FindAllStringSubmatch(env.GetCurrentDir(), -1)
	folderName := allString[0][1]

	if quiet {
		packageData.Name = folderName
		packageData.Version = "1.0.0"
		packageData.MainSrc = "./src"
	} else {
		packageData.Name = getParamOrDef("Package name ("+folderName+")", folderName)
		packageData.Homepage = getParamOrDef("Homepage", "")
		packageData.Version = getParamOrDef("Version (1.0.0)", "1.0.0")
		packageData.Description = getParamOrDef("Description", "")
		packageData.MainSrc = getParamOrDef("Source folder (./src)", "./src")
	}

	if err := pkgmanager.SavePackageCurrent(packageData); err != nil {
		msg.Die("Failed to save package: %v", err)
	}

	jsonData, err := json.MarshalIndent(packageData, "", "  ")
	if err != nil {
		msg.Die("Failed to marshal package: %v", err)
	}
	msg.Info("\n" + string(jsonData))
}

// getParamOrDef gets the parameter or default value
func getParamOrDef(msg string, def ...string) string {
	input := &pterm.DefaultInteractiveTextInput

	if len(def) > 0 {
		input = input.WithDefaultValue(def[0])
	}

	result, _ := input.Show(msg)

	return result
}

// printHead prints the head message
func printHead() {
	msg.Info(`
This utility will walk you through creating a boss.json file.
It only covers the most common items, and tries to guess sensible defaults.

Use 'boss install <pkg>' afterwards to install a package and
save it as a dependency in the boss.json file.

Press ^C at any time to quit.
`)
}
