package cmd

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

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

func doInitialization(quiet bool) {
	if !quiet {
		printHead()
	}

	packageData, err := models.LoadPackage(true)
	if err != nil && !os.IsNotExist(err) {
		msg.Die("Fail on open dependencies file: %s", err)
	}

	rxp := regexp.MustCompile(`^.+\` + string(filepath.Separator) + `([^\\]+)$`)

	allString := rxp.FindAllStringSubmatch(env.GetCurrentDir(), -1)
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

	json := packageData.Save()
	msg.Info("\n" + string(json))
}

func getParamOrDef(msg string, def ...string) string {
	input := &pterm.DefaultInteractiveTextInput

	if len(def) > 0 {
		input = input.WithDefaultValue(def[0])
	}

	result, _ := input.Show(msg)

	return result
}

func printHead() {
	msg.Info(`
This utility will walk you through creating a boss.json file.
It only covers the most common items, and tries to guess sensible defaults.

Use 'boss install <pkg>' afterwards to install a package and
save it as a dependency in the boss.json file.

Press ^C at any time to quit.
`)
}
