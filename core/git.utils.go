package core

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/git/crazy"
	"github.com/hashload/boss/models"
	"golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func Purge() {

}

func GetAllRepositoryes() {
	loadPackage, e := models.LoadPackage(false)
	if e != nil {
		panic(e)
	}
	deps := loadPackage.Dependencies.(map[string]interface{})
	for repo, ver := range deps {
		path := consts.FOLDER_DEPENDENCIES + string(filepath.Separator) + strings.Split(repo, "/")[1]
		var rep *git.Repository
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			rep, _ = git.PlainClone(path, false, &git.CloneOptions{
				URL:      "https://github.com/" + repo + ".git",
				Tags:     git.TagFollowing,
				Progress: os.Stdout,
			})
		} else {
			rep, _ := git.PlainOpen(path)
			rep.Fetch(&git.FetchOptions{
				RemoteName: "origin",
				Progress:   os.Stdout,
			})
		}
		worktree, _ := rep.Worktree()
		worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/tags/" + ver.(string)),
		})
		submodules, e := worktree.Submodules()
		if e == nil {
			submodules.Update(&git.SubmoduleUpdateOptions{
				Init:              true,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			})
		}
	}
}

func initCacheRepo() {
	//storage := memory.NewStorage()
	usr, e := user.Current()
	if e != nil {
		log.Fatal(e)
	}
	pem, e := ioutil.ReadFile(usr.HomeDir + "/.ssh/id_rsa")
	if e != nil {
		panic(e)
	}
	signer, e := ssh.ParsePrivateKey(pem)
	if e != nil {
		panic(e)
	}
	aith := &ssh2.PublicKeys{User: "git", Signer: signer}

	fs := osfs.New(consts.FOLDER_DEPENDENCIES + "/.cache")
	fsW := osfs.New(consts.FOLDER_DEPENDENCIES + "/.cacheC")
	st, _ := crazy.NewStorage(fs)

	_, err := os.Stat(consts.FOLDER_DEPENDENCIES + "/.cache")
	var repo *git.Repository
	if os.IsNotExist(err) {

		repo, e = git.Clone(st, fsW, &git.CloneOptions{
			URL:  "git@github.com:Snakeice/boss_repo.git",
			Auth: aith,
		})
		if e != nil {
			panic(e)
		}
	} else {
		repo, _ = git.Open(st, fs)
	}
	repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Force:      true,
	})
	worktree, _ := repo.Worktree()
	worktree.Checkout(&git.CheckoutOptions{
		Branch: "master",
		Force:  true,
	})

	worktree.Pull(nil)
}

func main() {
	initCacheRepo();
}
