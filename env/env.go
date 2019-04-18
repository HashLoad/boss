package env

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils/dcc32"
	"github.com/mitchellh/go-homedir"
	"os"
	"path/filepath"
	"strings"
)

var Global bool
var Internal = false
var GlobalConfiguration, _ = LoadConfiguration(GetBossHome())

func HashDelphiPath() string {
	hasher := md5.New()
	hasher.Write([]byte(strings.ToLower(GlobalConfiguration.DelphiPath)))
	hashString := hex.EncodeToString(hasher.Sum(nil))
	if Internal {
		hashString = consts.BossInternalDir + hashString
	}
	return hashString
}

func GetInternalGlobalDir() string {
	return filepath.Join(GetBossHome(), consts.FolderDependencies, consts.BossInternalDir+HashDelphiPath())

}

func getwd() string {
	if Global {
		return filepath.Join(GetBossHome(), consts.FolderDependencies, HashDelphiPath())
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

	cacheDir = filepath.FromSlash(cacheDir)
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
	return filepath.Join(GetBossHome(), consts.FolderDependencies, consts.BinFolder)
}

func GetDcc32Dir() string {
	if GlobalConfiguration.DelphiPath != "" {
		return GlobalConfiguration.DelphiPath
	}

	byCmd := dcc32.GetDcc32DirByCmd()
	if len(byCmd) > 0 {
		return byCmd[0]
	}

	return ""
}

func GetCurrentDelphiVersionFromRegisty() string {
	return dcc32.GetDelphiVersionNumberName(GlobalConfiguration.DelphiPath)
}
