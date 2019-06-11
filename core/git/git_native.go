package git

import (
	"fmt"
	"github.com/hashload/boss/core/paths"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/ldez/go-git-cmd-wrapper/clone"
	"github.com/ldez/go-git-cmd-wrapper/git"
	"github.com/ldez/go-git-cmd-wrapper/types"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func hasGitClient() bool {
	command := exec.Command("where", "git")
	_, err := command.Output()
	return err == nil
}

func GetWrapperFetch(dep models.Dependency) bool {
	if !hasGitClient() {
		return false
	}

	gitDir := filepath.Join(env.GetCacheDir(), dep.GetHashName())
	dirModule := strconv.Quote(filepath.Join(env.GetModulesDir(), dep.GetName()))

	if s, e := git.Fetch(
		clone.Directory(dirModule),
		func(g *types.Cmd) {
			g.AddOptions("--git-dir=" + gitDir)
			g.AddOptions("--all")
		},
	); e != nil {
		println(s)
		e.Error()
	}

	command := exec.Command("cmd", "/c cd "+dirModule+" && git --git-dir="+gitDir+" fetch --all")
	fmt.Printf("%v", command)

	if o, err := command.CombinedOutput(); err != nil {
		println(string(o))
		return false
	}
	return true
}

func GetWrapperClone(dep models.Dependency) bool {
	if !hasGitClient() {
		return false
	}

	paths.EnsureCacheDir(dep)
	dir := filepath.Join(env.GetCacheDir(), dep.GetHashName())
	dirModule := filepath.Join(env.GetModulesDir(), dep.GetName())

	_ = os.RemoveAll(dir)

	if _, err := git.Clone(
		clone.SeparateGitDir(dir),
		clone.Repository(dep.GetURL()),
		clone.Directory(dirModule),
	); err != nil {
		return false
	}
	return true
}
