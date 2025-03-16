package setup

import (
	"os"
	"path/filepath"
	"time"

	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/installer"
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils"
)

func updateVersion(newVersion int64) {
	env.GlobalConfiguration.ConfigVersion = newVersion
	env.GlobalConfiguration.SaveConfiguration()
}

func needUpdate(toVersion int64) bool {
	return env.GlobalConfiguration.ConfigVersion < toVersion
}

func executeUpdate(version int64, update func()) {
	if needUpdate(version) {
		msg.Debug("\t\tRunning update to version %d", version)
		update()
		updateVersion(version)
	} else {
		msg.Debug("\t\tUpdate to version %d already performed", version)
	}
}

func migration() {

	executeUpdate(1,
		func() {
			env.GlobalConfiguration.InternalRefreshRate = 5
		})

	executeUpdate(2, func() {
		OldPath := filepath.Join(env.GetBossHome(), consts.FolderDependencies, consts.BossInternalDirOld+env.HashDelphiPath())
		newPath := filepath.Join(env.GetBossHome(), consts.FolderDependencies, consts.BossInternalDir+env.HashDelphiPath())
		err := os.Rename(OldPath, newPath)
		if !os.IsNotExist(err) {
			utils.HandleError(err)
		}
	})

	executeUpdate(3, func() {
		env.GlobalConfiguration.GitEmbedded = true
	})

	executeUpdate(5, func() {
		env.Internal = false
		env.GlobalConfiguration.LastInternalUpdate = time.Now().AddDate(-1000, 0, 0)
		modulesDir := filepath.Join(env.GetBossHome(), consts.FolderDependencies, env.HashDelphiPath())
		if _, err := os.Stat(modulesDir); os.IsNotExist(err) {
			return
		}

		err := os.Remove(filepath.Join(modulesDir, consts.FilePackageLock))
		utils.HandleError(err)
		modules, err := models.LoadPackage(false)
		if err != nil {
			return
		}

		installer.GlobalInstall([]string{}, modules, false, false)
		env.Internal = true
	})

	executeUpdate(6, func() {
		err := os.RemoveAll(env.GetInternalGlobalDir())
		utils.HandleError(err)
	})
}
