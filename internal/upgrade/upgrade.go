package upgrade

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hashload/boss/internal/version"
	"github.com/hashload/boss/msg"
	"github.com/minio/selfupdate"
)

const (
	githubOrganization = "HashLoad"
	githubRepository   = "boss"
)

var (
	assetName = fmt.Sprintf("boss-%s-%s.zip", runtime.GOOS, runtime.GOARCH)
	zipFolder = fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
)

func BossUpgrade(preRelease bool) error {
	releases, err := getBossReleases()
	if err != nil {
		return fmt.Errorf("failed to get boss releases: %w", err)
	}

	release, err := findLatestRelease(releases, preRelease)
	if err != nil {
		return fmt.Errorf("failed to find latest boss release: %w", err)
	}

	asset, err := findAsset(release)
	if err != nil {
		return err
	} else if asset == nil {
		return fmt.Errorf("no asset found")
	}

	if *asset.Name == version.Get().Version {
		msg.Info("boss is already up to date")
		return nil
	}

	file, err := downloadAsset(asset)
	if err != nil {
		return err
	}

	defer file.Close()
	defer os.Remove(file.Name())

	buff, err := getAssetFromZip(file, assetName)
	if err != nil {
		return fmt.Errorf("failed to get asset from zip: %w", err)
	}

	err = apply(buff)
	if err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	} else {
		msg.Info("Update applied successfully to %s", *release.TagName)
		return nil
	}
}

func apply(buff []byte) error {

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath, _ := filepath.Abs(ex)

	return selfupdate.Apply(bytes.NewBuffer(buff), selfupdate.Options{
		OldSavePath: fmt.Sprintf("%s_bkp", exePath),
		TargetPath:  exePath,
	})

}
