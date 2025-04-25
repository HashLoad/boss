package upgrade

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"errors"

	"github.com/google/go-github/v69/github"
	"github.com/hashload/boss/pkg/progress"
	"github.com/snakeice/gogress"
)

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

func findLatestRelease(
	releases []*github.RepositoryRelease,
	preRelease bool,
) (*github.RepositoryRelease, error) {
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

func findAsset(
	release *github.RepositoryRelease,
	assetPrefix string,
) (*github.ReleaseAsset, error) {
	for _, asset := range release.Assets {
		if strings.HasPrefix(asset.GetName(), assetPrefix) {
			return asset, nil
		}
	}

	return nil, errors.New("no asset found")
}

func downloadAsset(asset *github.ReleaseAsset) (*os.File, error) {
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		asset.GetBrowserDownloadURL(),
		nil,
	)
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

	progress.Setup()
	bar := gogress.New64(int64(asset.GetSize()))

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
