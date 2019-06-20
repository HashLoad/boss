package core

import (
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func RunGC(ignoreLastUpdate bool) {
	_ = filepath.Walk(filepath.Join(env.GetCacheDir(), "info"), func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}

		var extension = filepath.Ext(info.Name())
		base := filepath.Base(info.Name())
		var name = strings.TrimRight(base, extension)
		repoInfo, e := models.RepoData(name)
		if e != nil {
			msg.Warn("Fail to parse repo info in GC: ", e)
			return nil
		}

		lastUpdate := repoInfo.LastUpdate.AddDate(0, 0, env.GlobalConfiguration.PurgeTime)
		if lastUpdate.Before(time.Now()) || ignoreLastUpdate {
			_ = os.RemoveAll(filepath.Join(env.GetCacheDir(), repoInfo.Key))
			_ = os.RemoveAll(filepath.Join(env.GetCacheDir(), "info", info.Name()))
		}

		return nil
	})
	env.GlobalConfiguration.LastPurge = time.Now()
	env.GlobalConfiguration.SaveConfiguration()
}
