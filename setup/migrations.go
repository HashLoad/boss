package setup

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"os"
	"path/filepath"
	"time"
)

func incVersion() {
	env.GlobalConfiguration.ConfigVersion++
	env.GlobalConfiguration.SaveConfiguration()
}

func needUpdate(toVersion int64) bool {
	return env.GlobalConfiguration.ConfigVersion < toVersion
}

func executeUpdate(version int64, update func()) {
	if needUpdate(version) {
		msg.Debug("\t\tRunning update to version %d", version)
		update()
		incVersion()
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

	executeUpdate(4, func() {
		env.GlobalConfiguration.LastInternalUpdate = time.Now().AddDate(-1000, 0, 0)
	})
}
