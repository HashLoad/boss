// Package gitadapter provides native Git command execution.
// This file implements Git clone/update operations using system Git commands.
package gitadapter

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	git2 "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
)

// gitCommand builds a system git command. On Windows it forces core.longpaths=true
// so checkouts of repositories with paths beyond MAX_PATH (260 chars) don't fail
// with "Filename too long" when the user's git config lacks that setting.
func gitCommand(args ...string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		args = append([]string{"-c", "core.longpaths=true"}, args...)
	}
	//nolint:gosec,nolintlint // Git command with controlled arguments
	return exec.CommandContext(context.Background(), "git", args...) // #nosec G204 -- Controlled git command
}

func checkHasGitClient() {
	command := exec.CommandContext(context.Background(), "where", "git")
	_, err := command.Output()
	if err != nil {
		msg.Die("❌ 'git.exe' not found in path")
	}
}

// CloneCacheNative clones the dependency repository to the cache using the native git client.
func CloneCacheNative(dep domain.Dependency) (*git2.Repository, error) {
	msg.Info("📥 Downloading dependency %s", dep.Repository)
	if err := doClone(dep); err != nil {
		return nil, err
	}
	return GetRepository(dep), nil
}

// UpdateCacheNative updates the dependency repository in the cache using the native git client.
func UpdateCacheNative(dep domain.Dependency) (*git2.Repository, error) {
	if err := getWrapperFetch(dep); err != nil {
		return nil, err
	}
	return GetRepository(dep), nil
}

func doClone(dep domain.Dependency) error {
	checkHasGitClient()

	dirModule := filepath.Join(env.GetModulesDir(), dep.Name())
	dir := "--separate-git-dir=" + filepath.Join(env.GetCacheDir(), dep.HashName())

	err := os.RemoveAll(dirModule)
	if err != nil && !os.IsNotExist(err) {
		msg.Debug("Failed to remove module directory: %v", err)
	}
	err = os.Remove(dirModule)
	if err != nil && !os.IsNotExist(err) {
		msg.Debug("Failed to remove module file: %v", err)
	}

	args := []string{"clone", dir}

	if env.GetGitShallow() {
		msg.Debug("Using shallow clone for %s", dep.Repository)
		args = append(args, "--depth", "1", "--single-branch")
	}

	args = append(args, dep.GetURL(), dirModule)

	cmd := gitCommand(args...)

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
		if err != nil {
			return fmt.Errorf("failed to create module directory: %w", err)
		}
	}

	writeDotGitFile(dep)
	cmdReset := gitCommand("reset", "--hard")
	cmdReset.Dir = dirModule
	if err := runCommand(cmdReset); err != nil {
		return err
	}

	cmd := gitCommand("fetch", "--all")
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
	cmd := gitCommand("submodule", "update", "--init", "--recursive")
	cmd.Dir = dirModule

	if err := runCommand(cmd); err != nil {
		return err
	}
	return nil
}

// CheckoutNative switches the dependency repository to the given reference using system git.
func CheckoutNative(dep domain.Dependency, referenceName plumbing.ReferenceName) error {
	dirModule := filepath.Join(env.GetModulesDir(), dep.Name())
	cmd := gitCommand("checkout", "-f", referenceName.Short())
	cmd.Dir = dirModule
	return runCommand(cmd)
}

// PullNative fetches and merges updates using system git.
func PullNative(dep domain.Dependency) error {
	dirModule := filepath.Join(env.GetModulesDir(), dep.Name())
	cmd := gitCommand("pull", "--force")
	cmd.Dir = dirModule
	return runCommand(cmd)
}

func runCommand(cmd *exec.Cmd) error {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	cmd.Env = os.Environ()

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
