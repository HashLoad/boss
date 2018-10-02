package env

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/msg"
	"github.com/mitchellh/go-homedir"
	"os"
	"path/filepath"
)

func GetCacheDir() string {
	s := os.Getenv("BOSS_CACHE_DIR")

	if s == "" {
		s = filepath.Join(GetBossHome(), "cache")
	}
	return s
}

func GetBossHome() string {
	caheDir, e := homedir.Dir()
	if e != nil {
		msg.Err("Error to get cache paths", e)
	}
	return filepath.Join(caheDir, ".boss")
}

func GetModulesDir() string {
	dir, err := os.Getwd()
	if err != nil {
		msg.Err("Error to get module paths", err)
	}
	return filepath.Join(dir, consts.FOLDER_DEPENDENCIES)
}
