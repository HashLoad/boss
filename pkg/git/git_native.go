package git

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	git2 "github.com/go-git/go-git/v5"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/pkg/paths"
	"github.com/hashload/boss/utils"
)

func checkHasGitClient() {
	command := exec.Command("where", "git")
	_, err := command.Output()
	if err != nil {
		msg.Die("Git.exe not found in path")
	}
}

func CloneCacheNative(dep models.Dependency) *git2.Repository {
	msg.Info("Downloading dependency %s", dep.Repository)
	doClone(dep)
	return GetRepository(dep)
}

func UpdateCacheNative(dep models.Dependency) *git2.Repository {
	getWrapperFetch(dep)
	return GetRepository(dep)
}

func doClone(dep models.Dependency) {
	checkHasGitClient()

	paths.EnsureCacheDir(dep)

	dirModule := filepath.Join(env.GetModulesDir(), dep.Name())
	dir := "--separate-git-dir=" + filepath.Join(env.GetCacheDir(), dep.HashName())

	err := os.RemoveAll(dirModule)
	if !os.IsNotExist(err) {
		utils.HandleError(err)
	}
	err = os.Remove(dirModule)
	if !os.IsNotExist(err) {
		utils.HandleError(err)
	}

	cmd := exec.Command("git", "clone", dir, dep.GetURL(), dirModule)

	if err = runCommand(cmd); err != nil {
		msg.Die(err.Error())
	}
	initSubmodulesNative(dep)

	_ = os.Remove(filepath.Join(dirModule, ".git"))
}

func writeDotGitFile(dep models.Dependency) {
	mask := fmt.Sprintf("gitdir: %s\n", filepath.Join(env.GetCacheDir(), dep.HashName()))
	path := filepath.Join(env.GetModulesDir(), dep.Name(), ".git")
	_ = os.WriteFile(path, []byte(mask), 0600)
}

func getWrapperFetch(dep models.Dependency) {
	checkHasGitClient()

	dirModule := filepath.Join(env.GetModulesDir(), dep.Name())

	if _, err := os.Stat(dirModule); os.IsNotExist(err) {
		err = os.MkdirAll(dirModule, 0600)
		utils.HandleError(err)
	}

	writeDotGitFile(dep)
	cmdReset := exec.Command("git", "reset", "--hard")
	cmdReset.Dir = dirModule
	if err := runCommand(cmdReset); err != nil {
		msg.Die(err.Error())
	}

	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = dirModule

	if err := runCommand(cmd); err != nil {
		msg.Die(err.Error())
	}

	initSubmodulesNative(dep)

	_ = os.Remove(filepath.Join(dirModule, ".git"))
}

func initSubmodulesNative(dep models.Dependency) {
	dirModule := filepath.Join(env.GetModulesDir(), dep.Name())
	cmd := exec.Command("git", "submodule", "update", "--init", "--recursive")
	cmd.Dir = dirModule

	if err := runCommand(cmd); err != nil {
		msg.Die(err.Error())
	}
}

func runCommand(cmd *exec.Cmd) error {
	cmd.Stdout = newWriter(false)
	cmd.Stderr = newWriter(true)
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

type writer struct {
	io.Writer
	errorWritter bool
}

func newWriter(errorWritter bool) *writer {
	return &writer{errorWritter: errorWritter}
}

func (writer *writer) Write(p []byte) (int, error) {
	var str = "  " + string(p)
	if writer.errorWritter {
		msg.Err(str)
	} else {
		msg.Info(str)
	}
	return len(p), nil
}
