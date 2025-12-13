package env

import (
	//nolint:gosec // We are not using this for security purposes
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils/dcc32"
	"github.com/mitchellh/go-homedir"
)

//nolint:gochecknoglobals //TODO: Refactor this
var (
	global                 bool
	internal               = false
	globalConfiguration, _ = LoadConfiguration(GetBossHome())
)

// SetGlobal sets the global flag
func SetGlobal(b bool) {
	global = b
}

// SetInternal sets the internal flag
func SetInternal(b bool) {
	internal = b
}

// GetInternal returns the internal flag
func GetInternal() bool {
	return internal
}

// GetGlobal returns the global flag
func GetGlobal() bool {
	return global
}

// GlobalConfiguration returns the global configuration
func GlobalConfiguration() *Configuration {
	return globalConfiguration
}

// HashDelphiPath returns the hash of the Delphi path
func HashDelphiPath() string {
	//nolint:gosec // We are not using this for security purposes
	hasher := md5.New()
	hasher.Write([]byte(strings.ToLower(GlobalConfiguration().DelphiPath)))
	hashString := hex.EncodeToString(hasher.Sum(nil))
	if internal {
		hashString = consts.BossInternalDir + hashString
	}
	return hashString
}

// GetInternalGlobalDir returns the internal global directory
func GetInternalGlobalDir() string {
	internalOld := internal
	internal = true
	result := filepath.Join(GetBossHome(), consts.FolderDependencies, HashDelphiPath())
	internal = internalOld
	return result
}

// getwd returns the working directory
func getwd() string {
	if global {
		return filepath.Join(GetBossHome(), consts.FolderDependencies, HashDelphiPath())
	}

	dir, err := os.Getwd()
	if err != nil {
		msg.Err("Error to get paths", err)
		return ""
	}

	return dir
}

// GetCacheDir returns the cache directory
func GetCacheDir() string {
	return filepath.Join(GetBossHome(), "cache")
}

// GetBossHome returns the Boss home directory
func GetBossHome() string {
	homeDir := os.Getenv("BOSS_HOME")

	if homeDir == "" {
		systemHome, err := homedir.Dir()
		homeDir = systemHome
		if err != nil {
			msg.Err("Error to get cache paths", err)
		}

		homeDir = filepath.FromSlash(homeDir)
	}
	return filepath.Join(homeDir, consts.FolderBossHome)
}

// GetBossFile returns the Boss file path
func GetBossFile() string {
	return filepath.Join(GetCurrentDir(), consts.FilePackage)
}

// GetModulesDir returns the modules directory
func GetModulesDir() string {
	return filepath.Join(GetCurrentDir(), consts.FolderDependencies)
}

// GetCurrentDir returns the current directory
func GetCurrentDir() string {
	return getwd()
}

// GetGlobalEnvBpl returns the global environment BPL directory
func GetGlobalEnvBpl() string {
	return filepath.Join(GetBossHome(), consts.FolderEnvBpl)
}

// GetGlobalEnvDcp returns the global environment DCP directory
func GetGlobalEnvDcp() string {
	return filepath.Join(GetBossHome(), consts.FolderEnvDcp)
}

// GetGlobalEnvDcu returns the global environment DCU directory
func GetGlobalEnvDcu() string {
	return filepath.Join(GetBossHome(), consts.FolderEnvDcu)
}

// GetGlobalBinPath returns the global binary path
func GetGlobalBinPath() string {
	return filepath.Join(GetBossHome(), consts.FolderDependencies, consts.BinFolder)
}

// GetDcc32Dir returns the DCC32 directory
func GetDcc32Dir() string {
	if GlobalConfiguration().DelphiPath != "" {
		return GlobalConfiguration().DelphiPath
	}

	byCmd := dcc32.GetDcc32DirByCmd()
	if len(byCmd) > 0 {
		return byCmd[0]
	}

	return ""
}
