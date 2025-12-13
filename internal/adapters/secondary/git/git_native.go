package gitadapter

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	git2 "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/paths"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils"
)

func checkHasGitClient() {
	command := exec.Command("where", "git")
	_, err := command.Output()
	if err != nil {
		msg.Die("Git.exe not found in path")
	}
}

func CloneCacheNative(dep domain.Dependency) (*git2.Repository, error) {
	msg.Info("Downloading dependency %s", dep.Repository)
	if err := doClone(dep); err != nil {
		return nil, err
	}
	return GetRepository(dep), nil
}

func UpdateCacheNative(dep domain.Dependency) (*git2.Repository, error) {
	if err := getWrapperFetch(dep); err != nil {
		return nil, err
	}
	return GetRepository(dep), nil
}

func doClone(dep domain.Dependency) error {
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
		return err
	}
	if err := initSubmodulesNative(dep); err != nil {
		return err
	}

	_ = os.Remove(filepath.Join(dirModule, ".git"))
	return nil
}

func writeDotGitFile(dep domain.Dependency) {
	mask := fmt.Sprintf("gitdir: %s\n", filepath.Join(env.GetCacheDir(), dep.HashName()))
	path := filepath.Join(env.GetModulesDir(), dep.Name(), ".git")
	_ = os.WriteFile(path, []byte(mask), 0600)
}

func getWrapperFetch(dep domain.Dependency) error {
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
		return err
	}

	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = dirModule

	if err := runCommand(cmd); err != nil {
		return err
	}

	if err := initSubmodulesNative(dep); err != nil {
		return err
	}

	_ = os.Remove(filepath.Join(dirModule, ".git"))
	return nil
}

func initSubmodulesNative(dep domain.Dependency) error {
	dirModule := filepath.Join(env.GetModulesDir(), dep.Name())
	cmd := exec.Command("git", "submodule", "update", "--init", "--recursive")
	cmd.Dir = dirModule

	if err := runCommand(cmd); err != nil {
		return err
	}
	return nil
}

func CheckoutNative(dep domain.Dependency, referenceName plumbing.ReferenceName) error {
	dirModule := filepath.Join(env.GetModulesDir(), dep.Name())
	cmd := exec.Command("git", "checkout", "-f", referenceName.Short())
	cmd.Dir = dirModule
	return runCommand(cmd)
}

func PullNative(dep domain.Dependency) error {
	dirModule := filepath.Join(env.GetModulesDir(), dep.Name())
	cmd := exec.Command("git", "pull", "--force")
	cmd.Dir = dirModule
	return runCommand(cmd)
}

func runCommand(cmd *exec.Cmd) error {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %w\nStderr: %s", err, stderrBuf.String())
	}

	if stdoutBuf.Len() > 0 {
		msg.Debug("Command stdout: %s", stdoutBuf.String())
	}
	if stderrBuf.Len() > 0 {
		msg.Debug("Command stderr: %s", stderrBuf.String())
	}

	return nil
}
