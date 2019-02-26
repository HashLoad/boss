package compiler

import (
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func isCommandAvailable(name string) bool {
	cmd := exec.Command(name, "-version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

//noinspection GoUnhandledErrorResult
func BuildDucs() {

	paths := getNewPaths(filepath.Dir("./modules"))
	additionalPaths := "-U" + strconv.Quote(paths)

	if !isCommandAvailable("dcc32.exe") {
		msg.Warn("dcc32 not found in path")
		return
	}

	_ = filepath.Walk("./modules",
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			if filepath.Ext(path) != ".pas" {
				return nil
			}
			command := exec.Command("dcc32.exe", filepath.Base(path), additionalPaths)
			command.Dir = filepath.Dir(path)
			command.Output()
			return nil
		})

	_ = filepath.Walk("./modules",
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			if filepath.Ext(path) != ".dpk" {
				return nil
			}
			msg.Print("running dcc32.exe " + filepath.Base(path))
			command := exec.Command("dcc32.exe", filepath.Base(path), additionalPaths)
			command.Dir = filepath.Dir(path)
			command.Output()
			return nil
		})
}

func getNewPaths(path string) string {
	paths := []string{}
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		matched, _ := regexp.MatchString(".*.pas$", info.Name())
		dir := filepath.Dir(path)
		dir, _ = filepath.Abs(dir)
		if matched && !utils.Contains(paths, dir) {
			paths = append(paths, dir)
		}
		return nil
	})
	return strings.Join(paths, ";")
}
