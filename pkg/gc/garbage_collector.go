package gc

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
)

func CleanupCache(ignoreLastUpdate, allFiles bool) error {
	defer func() {
		env.GlobalConfiguration().LastPurge = time.Now()
		env.GlobalConfiguration().SaveConfiguration()
	}()

	path := filepath.Join(env.GetCacheDir(), "info")
	return filepath.Walk(path, deleteCacheArtifacts(ignoreLastUpdate, allFiles))
}

func isCleanable(repoInfo *models.RepoInfo, ignoreLastUpdate bool) bool {
	purgeDeadline := repoInfo.LastUpdate.AddDate(0, 0, env.GlobalConfiguration().PurgeTime)
	if purgeDeadline.After(time.Now()) && !ignoreLastUpdate {
		return false
	}

	for _, ref := range repoInfo.References {
		_, err := os.Stat(ref)
		if os.IsNotExist(err) {
			continue
		}

		return false
	}

	return true
}

func deleteCacheArtifacts(ignoreLastUpdate, allFiles bool) filepath.WalkFunc {
	return func(_ string, info os.FileInfo, _ error) error {
		if info == nil || info.IsDir() {
			return nil
		}

		var extension = filepath.Ext(info.Name())
		base := filepath.Base(info.Name())
		var name = strings.TrimRight(base, extension)
		repoInfo, err := models.RepoData(name)
		if err != nil {
			msg.Warn("Fail to parse repo info in GC: ", err)
			return nil
		}

		if !isCleanable(repoInfo, ignoreLastUpdate) && !allFiles {
			return nil
		}

		msg.Info("Cleaning %s", repoInfo.Name)

		_ = os.RemoveAll(filepath.Join(env.GetCacheDir(), repoInfo.Key))
		_ = os.RemoveAll(filepath.Join(env.GetCacheDir(), repoInfo.Key+"_wt"))
		_ = os.RemoveAll(filepath.Join(env.GetCacheDir(), "info", info.Name()))

		return nil
	}
}
