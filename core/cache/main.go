package cache

import (
	"io"
	"os"
	"path/filepath"

	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
)

func HasCache(dep models.Dependency) bool {
	dir := filepath.Join(env.GetCacheDir(), dep.GetHashName())
	info, err := os.Stat(dir)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	if !info.IsDir() {
		os.RemoveAll(dir)
		return false
	}
	return dirIsEmpty(dir)

}

func dirIsEmpty(dir string) bool {
	f, err := os.Open(dir)
	if err != nil {
		return false
	}
	defer f.Close()
	_, err = f.Readdir(1)
	if err == io.EOF {
		return true
	}
	return false
}
