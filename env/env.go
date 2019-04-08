package env

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/msg"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/sys/windows/registry"
	"os"
	"os/exec"
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

	command := exec.Command("where", "dcc32")
	output, err := command.Output()
	if err != nil {
		msg.Warn("dcc32 not found")
	}
	outputStr := strings.ReplaceAll(string(output), "\n", "")
	outputStr = strings.ReplaceAll(outputStr, "\r", "")
	outputStr = filepath.Dir(outputStr)

	return outputStr
}

func HandleError(err error) {
	if err != nil {
		msg.Err(err.Error())
	}
}

func HandleErrorFatal(err error) {
	if err != nil {
		msg.Die(err.Error())
	}
}

func GetDelphiVersionFromRegisty() string {

	delphiVersions, err := registry.OpenKey(registry.CURRENT_USER, `Software\Embarcadero\BDS\`,
		registry.ALL_ACCESS)
	if err != nil {
		msg.Err("Cannot open registry to IDE version")
	}
	keyInfo, err := delphiVersions.Stat()
	HandleError(err)

	names, err := delphiVersions.ReadSubKeyNames(int(keyInfo.SubKeyCount))
	HandleError(err)

	for _, value := range names {
		delphiInfo, err := registry.OpenKey(registry.CURRENT_USER, `Software\Embarcadero\BDS\`+value, registry.QUERY_VALUE)
		HandleError(err)

		appPath, _, err := delphiInfo.GetStringValue("App")
		if os.IsNotExist(err) {
			continue
		}
		HandleError(err)
		if strings.HasPrefix(strings.ToLower(appPath), strings.ToLower(GlobalConfiguration.DelphiPath)) {
			return value
		}

	}
	return ""
}
