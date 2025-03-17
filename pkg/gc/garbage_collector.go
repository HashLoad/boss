package gc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
)

func RunGC(ignoreLastUpdate bool) error {
	defer func() {
		env.GlobalConfiguration().LastPurge = time.Now()
		env.GlobalConfiguration().SaveConfiguration()
	}()

	return filepath.Walk(filepath.Join(env.GetCacheDir(), "info"),
		func(_ string, info os.FileInfo, _ error) error {
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

			lastUpdate := repoInfo.LastUpdate.AddDate(0, 0, env.GlobalConfiguration().PurgeTime)
			if lastUpdate.Before(time.Now()) || ignoreLastUpdate {
				_ = os.RemoveAll(filepath.Join(env.GetCacheDir(), repoInfo.Key))
				_ = os.RemoveAll(filepath.Join(env.GetCacheDir(), fmt.Sprintf("%s_wt", repoInfo.Key)))
				_ = os.RemoveAll(filepath.Join(env.GetCacheDir(), "info", info.Name()))
			}

			return nil
		})
}
