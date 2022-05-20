package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/spf13/cobra"
)

var quiet bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project",
	Long:  "Initialize a new project and creates a boss.json file",
	Example: `  Initialize a new project:
  boss init

  Initialize a new project without having it ask any questions:
  boss init --quiet`,
	Run: func(cmd *cobra.Command, args []string) {
		doInitialization(quiet)
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "without asking questions")
}

func doInitialization(quiet bool) {
	if !quiet {
		printHead()
	}

	pkgJson, _ := models.LoadPackage(true)

	rxp := regexp.MustCompile(`^.+\` + string(filepath.Separator) + `([^\\]+)$`)

	allString := rxp.FindAllStringSubmatch(env.GetCurrentDir(), -1)
	folderName := allString[0][1]

	if quiet {
		pkgJson.Name = folderName
		pkgJson.Version = "1.0.0"
		pkgJson.MainSrc = "./"
	} else {
		pkgJson.Name = getParamOrDef("Package name ("+folderName+")", folderName)
		pkgJson.Homepage = getParamOrDef("Homepage", "")
		pkgJson.Version = getParamOrDef("Version (1.0.0)", "1.0.0")
		pkgJson.Description = getParamOrDef("Description", "")
		pkgJson.MainSrc = getParamOrDef("Source folder (./src)", "./src")
	}

	json := pkgJson.Save()
	fmt.Println("\n" + string([]byte(json)))
}

func getParamOrDef(msg string, def string) string {
	fmt.Print(msg + ": ")
	rr := bufio.NewReader(os.Stdin)
	if res, e := rr.ReadString('\n'); e == nil && res != "\n" {
		res = strings.ReplaceAll(res, "\t", "")
		res = strings.ReplaceAll(res, "\n", "")
		res = strings.ReplaceAll(res, "\r", "")
		if res == "" {
			return def
		} else {
			return res
		}
	}
	return def
}

func printHead() {
	println(`
This utility will walk you through creating a boss.json file.
It only covers the most common items, and tries to guess sensible defaults.

Use 'boss install <pkg>' afterwards to install a package and
save it as a dependency in the boss.json file.

Press ^C at any time to quit.
`)
}
