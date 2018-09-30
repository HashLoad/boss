package cache

import (
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"os"
	"path/filepath"
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
	return info.IsDir()
}
