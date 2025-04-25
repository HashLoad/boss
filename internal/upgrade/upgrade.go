package upgrade

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hashload/boss/internal/version"
	"github.com/hashload/boss/pkg/msg"
	"github.com/minio/selfupdate"
)

const (
	githubOrganization = "HashLoad"
	githubRepository   = "boss"
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

	assetPrefix := assetNameWithouExt()

	asset, err := findAsset(release, assetPrefix)
	if err != nil {
		return err
	}

	if asset == nil {
		return errors.New("no asset found")
	}

	if asset.GetName() == version.Get().Version {
		msg.Info("boss is already up to date")
		return nil
	}

	msg.Info("Downloading update %s -> %s", version.Get().Version, release.GetTagName())

	file, err := downloadAsset(asset)
	if err != nil {
		return err
	}

	defer file.Close()
	defer os.Remove(file.Name())

	buff, err := getAssetFromFile(file, asset)
	if err != nil {
		return fmt.Errorf("failed to get asset from compressed file: %w", err)
	}

	err = apply(buff)
	if err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	msg.Info("Update applied successfully %s -> %s", version.Get().Version, release.GetTagName())
	return nil
}

func apply(buff []byte) error {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath, _ := filepath.Abs(ex)

	return selfupdate.Apply(bytes.NewBuffer(buff), selfupdate.Options{
		OldSavePath: exePath + "_bkp",
		TargetPath:  exePath,
	})
}

func assetNameWithouExt() string {
	return fmt.Sprintf("boss-%s-%s", runtime.GOOS, runtime.GOARCH)
}
