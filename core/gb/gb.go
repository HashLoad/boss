package gb

import (
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func RunGB() {
	msg.Info("Running GB...")
	filepath.Walk(filepath.Join(env.GetCacheDir(), "info"), func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		var extension = filepath.Ext(info.Name())
		base := filepath.Base(info.Name())
		var name = strings.TrimRight(base, extension)
		repoInfo, e := models.RepoData(name)
		if e != nil {
			msg.Warn("Fail to parse repo info in GB: ", e)
			return nil
		}

		lastUpdate := repoInfo.LastUpdate.AddDate(0, 0, models.GlobalConfiguration.PurgeTime)
		if lastUpdate.Before(time.Now()) {
			os.Remove(filepath.Join(env.GetCacheDir(), repoInfo.Key))
			os.Remove(filepath.Join(env.GetCacheDir(), "info", info.Name()))
		}

		return nil
	})
	models.GlobalConfiguration.LastPurge = time.Now().String()
	models.GlobalConfiguration.SaveConfiguration()
}
