package core

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
)

func printHead() {
	println(`
This utility will walk you through creating a boss.json file.
It only covers the most common items, and tries to guess sensible defaults.
		 
Use 'boss install <pkg>' afterwards to install a package and
save it as a dependency in the boss.json file.

Press ^C at any time to quit.`)
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

func doInitialization(quiet bool) {
	if !quiet {
		printHead()
	}

	pkgJson, _ := models.LoadPackage(true)

	var folderName = ""
	rxp, err := regexp.Compile(`^.+\` + string(filepath.Separator) + `([^\\]+)$`)
	if err == nil {
		allString := rxp.FindAllStringSubmatch(env.GetCurrentDir(), -1)
		folderName = allString[0][1]
	}

	if quiet {
		pkgJson.Name = folderName
		pkgJson.Version = "1.0.0"
		pkgJson.MainSrc = "./"
	} else {
		pkgJson.Name = getParamOrDef("package name: ("+folderName+")", folderName)
		pkgJson.Homepage = getParamOrDef("homepage", "")
		pkgJson.Version = getParamOrDef("version: (1.0.0)", "1.0.0")
		pkgJson.Description = getParamOrDef("description", "")
		pkgJson.MainSrc = getParamOrDef("source folder: (./)", "./")
	}

	pkgJson.Save()
}

func InitializeBossPackage(quiet bool) {
	doInitialization(quiet)
}
