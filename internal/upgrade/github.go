// Package upgrade provides GitHub API integration for fetching Boss releases.
// This file handles release discovery, asset filtering, and download.
package upgrade

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"

	"errors"

	"github.com/google/go-github/v69/github"
	"github.com/snakeice/gogress"
)

// getBossReleases returns the boss releases
func getBossReleases() ([]*github.RepositoryRelease, error) {
	gh := github.NewClient(nil)

	releases := []*github.RepositoryRelease{}
	page := 0
	for {
		listOptions := github.ListOptions{
			Page:    page,
			PerPage: 20,
		}

		releasesPage, resp, err := gh.Repositories.ListReleases(
			context.Background(),
			githubOrganization,
			githubRepository,
			&listOptions,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to get releases: %w", err)
		}

		releases = append(releases, releasesPage...)

		if resp.NextPage == 0 {
			break
		}

		page = resp.NextPage
	}

	return releases, nil
}

// findLatestRelease finds the latest release
func findLatestRelease(releases []*github.RepositoryRelease, preRelease bool) (*github.RepositoryRelease, error) {
	var bestRelease *github.RepositoryRelease

	for _, release := range releases {
		if release.GetPrerelease() && !preRelease {
			continue
		}

		if bestRelease == nil || release.GetTagName() > bestRelease.GetTagName() {
			bestRelease = release
		}
	}

	if bestRelease == nil {
		return nil, errors.New("no releases found")
	}

	return bestRelease, nil
}

// findAsset finds the asset in the release
func findAsset(release *github.RepositoryRelease) (*github.ReleaseAsset, error) {
	for _, asset := range release.Assets {
		if asset.GetName() == getAssetName() {
			return asset, nil
		}
	}

	return nil, errors.New("no asset found")
}

// downloadAsset downloads the asset
func downloadAsset(asset *github.ReleaseAsset) (*os.File, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, asset.GetBrowserDownloadURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	file, err := os.CreateTemp("", "boss")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	bar := gogress.New64(int64(math.Round(float64(asset.GetSize()))))
	bar.Start()
	defer bar.Finish()
	proxyReader := bar.NewProxyReader(resp.Body)
	defer proxyReader.Close()

	_, err = io.Copy(file, proxyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy asset: %w", err)
	}

	return file, nil
}
