package upgrade

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hashload/boss/internal/version"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/msg"
	"github.com/minio/selfupdate"
)

const (
	githubOrganization = "HashLoad"
	githubRepository   = "boss"
)

// BossUpgrade performs the self-update of the boss executable.
// It checks for the latest release on GitHub, downloads it, and applies the update.
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
		return errors.New("no asset found")
	}

	if *asset.Name == version.Get().Version {
		msg.Info(consts.StatusMsgAlreadyUpToDate)
		return nil
	}

	file, err := downloadAsset(asset)
	if err != nil {
		return err
	}

	defer file.Close()
	defer os.Remove(file.Name())

	buff, err := getAssetFromFile(file, getAssetName())
	if err != nil {
		return fmt.Errorf("failed to get asset from zip: %w", err)
	}

	err = apply(buff)
	if err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	msg.Info("Update applied successfully to %s", *release.TagName)
	return nil
}

// apply applies the update
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

// getAssetName returns the asset name
func getAssetName() string {
	ext := "zip"
	if runtime.GOOS != "windows" {
		ext = "tar.gz"
	}

	return fmt.Sprintf("boss-%s-%s.%s", runtime.GOOS, runtime.GOARCH, ext)
}
