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
	internalOld := Internal
	Internal = true
	result := filepath.Join(GetBossHome(), consts.FolderDependencies, HashDelphiPath())
	Internal = internalOld
	return result
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
	return filepath.Join(GetBossHome(), "cache")
}

func GetBossHome() string {

	homeDir := os.Getenv("BOSS_HOME")

	if homeDir == "" {
		systemHome, e := homedir.Dir()
		homeDir = systemHome
		if e != nil {
			msg.Err("Error to get cache paths", e)
		}

		homeDir = filepath.FromSlash(homeDir)
	}
	return filepath.Join(homeDir, consts.FolderBossHome)
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

func GetGlobalEnvBpl() string {
	return filepath.Join(GetBossHome(), consts.FolderEnvBpl)
}
func GetGlobalEnvDcp() string {
	return filepath.Join(GetBossHome(), consts.FolderEnvDcp)
}
func GetGlobalEnvDcu() string {
	return filepath.Join(GetBossHome(), consts.FolderEnvDcu)
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

func GetCurrentDelphiVersionFromRegistry() string {
	return dcc32.GetDelphiVersionNumberName(GlobalConfiguration.DelphiPath)
}
