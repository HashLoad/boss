// Package setup handles application initialization, migrations, and environment configuration.
// It creates necessary directories, runs database migrations, and initializes the Delphi environment.
package setup

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	filesystem "github.com/hashload/boss/internal/adapters/secondary/filesystem"
	registry "github.com/hashload/boss/internal/adapters/secondary/registry"
	"github.com/hashload/boss/internal/adapters/secondary/repository"
	"github.com/hashload/boss/internal/core/services/installer"
	"github.com/hashload/boss/internal/core/services/packages"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/pkg/pkgmanager"
	"github.com/hashload/boss/utils/dcc32"
)

// PATH is the environment variable for the system path.
const PATH string = "PATH"

// DefaultModules returns the list of default internal modules.
func DefaultModules() []string {
	return []string{
		"bpl-identifier",
	}
}

// Initialize initializes the Boss environment.
func Initialize() {
	initializeInfrastructure()

	var oldGlobal = env.GetGlobal()
	env.SetInternal(true)
	env.SetGlobal(true)

	msg.Debug("DEBUG MODE")
	msg.Debug("\tInitializing delphi version")
	initializeDelphiVersion()

	msg.Debug("\tExecuting migrations")
	migration()
	msg.Debug("\tInstalling internal modules")
	installModules(DefaultModules())
	msg.Debug("\tCreating paths")
	CreatePaths()

	InitializePath()

	env.SetGlobal(oldGlobal)
	env.SetInternal(false)
	msg.Debug("finish boss system initialization")
}

// initializeInfrastructure sets up infrastructure dependencies.
// This is the composition root where we wire up adapters to ports.
func initializeInfrastructure() {
	fs := filesystem.NewOSFileSystem()
	packageRepo := repository.NewFilePackageRepository(fs)
	lockRepo := repository.NewFileLockRepository(fs)
	packageService := packages.NewPackageService(packageRepo, lockRepo)
	pkgmanager.SetInstance(packageService)
}

// CreatePaths creates the necessary paths for boss.
func CreatePaths() {
	_, err := os.Stat(env.GetGlobalEnvBpl())
	if os.IsNotExist(err) {
		_ = os.MkdirAll(env.GetGlobalEnvBpl(), 0755) // #nosec G301 -- Standard permissions for shared directory
	}
}

// installModules installs the internal modules.
func installModules(modules []string) {
	pkg, _ := pkgmanager.LoadPackage()
	encountered := 0
	for _, newPackage := range modules {
		for installed := range pkg.Dependencies {
			if strings.Contains(installed, newPackage) {
				encountered++
			}
		}
	}

	nextUpdate := env.GlobalConfiguration().LastInternalUpdate.
		AddDate(0, 0, env.GlobalConfiguration().PurgeTime)

	if encountered == len(modules) && time.Now().Before(nextUpdate) {
		return
	}

	env.GlobalConfiguration().LastInternalUpdate = time.Now()
	env.GlobalConfiguration().SaveConfiguration()

	installer.GlobalInstall(env.GlobalConfiguration(), modules, pkg, false, false)
	moveBptIdentifier()
}

// moveBptIdentifier moves the bpl identifier.
func moveBptIdentifier() {
	var outExeCompilation = filepath.Join(env.GetGlobalBinPath(), consts.BplIdentifierName)
	if _, err := os.Stat(outExeCompilation); os.IsNotExist(err) {
		return
	}

	var exePath = filepath.Join(env.GetModulesDir(), consts.BinFolder, consts.BplIdentifierName)
	err := os.MkdirAll(filepath.Dir(exePath), 0600)
	if err != nil {
		msg.Err("❌ %s", err.Error())
	}

	err = os.Rename(outExeCompilation, exePath)
	if err != nil {
		msg.Err("❌ %s", err.Error())
	}
}

// initializeDelphiVersion initializes the delphi version.
func initializeDelphiVersion() {
	if len(env.GlobalConfiguration().DelphiPath) != 0 {
		return
	}
	dcc32DirByCmd := dcc32.GetDcc32DirByCmd()
	if len(dcc32DirByCmd) != 0 {
		env.GlobalConfiguration().DelphiPath = dcc32DirByCmd[0]
		env.GlobalConfiguration().SaveConfiguration()
		return
	}

	byRegistry := registry.GetDelphiPaths()
	if len(byRegistry) != 0 {
		env.GlobalConfiguration().DelphiPath = byRegistry[len(byRegistry)-1]
		env.GlobalConfiguration().SaveConfiguration()
		return
	}
}
