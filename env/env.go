package env

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/msg"
	"github.com/mitchellh/go-homedir"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var Global bool
var GlobalConfiguration, _ = LoadConfiguration(GetBossHome())

func hashDelphiPath() string {
	hasher := md5.New()
	hasher.Write([]byte(strings.ToLower(GlobalConfiguration.DelphiPath)))
	return hex.EncodeToString(hasher.Sum(nil))
}

func getwd() string {
	if Global {
		return filepath.Join(GetBossHome(), consts.FolderDependencies, hashDelphiPath())
	} else {
		if dir, err := os.Getwd(); err != nil {
			msg.Err("Error to get paths", err)
			return ""
		} else {
			return dir
		}
	}
}

func GetCacheDir() string {
	s := os.Getenv("BOSS_CACHE_DIR")

	if s == "" {
		s = filepath.Join(GetBossHome(), "cache")
	}
	return s
}

func GetBossHome() string {
	cacheDir, e := homedir.Dir()
	if e != nil {
		msg.Err("Error to get cache paths", e)
	}
	return filepath.Join(cacheDir, consts.FolderBossHome)
}

func GetBossFile() string {
	return filepath.Join(GetCurrentDir(), consts.FilePackage)
}

func GetModulesDir() string {
	return filepath.Join(GetCurrentDir(), consts.FolderDependencies)
}

func GetCurrentDir() string {
	return getwd()
}

func GetGlobalBinPath() string {
	return path.Join(GetBossHome(), consts.FolderDependencies, consts.BinFolder)
}
