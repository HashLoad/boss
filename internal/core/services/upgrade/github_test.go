//nolint:testpackage // Testing internal functions
package upgrade

import (
	"testing"

	"github.com/google/go-github/v69/github"
)

// TestFindLatestRelease_NoReleases tests error when no releases available.
func TestFindLatestRelease_NoReleases(t *testing.T) {
	releases := []*github.RepositoryRelease{}

	_, err := findLatestRelease(releases, false)
	if err == nil {
		t.Error("findLatestRelease() should return error for empty releases")
	}
}

// TestFindLatestRelease_OnlyPreReleases tests filtering of prereleases.
func TestFindLatestRelease_OnlyPreReleases(t *testing.T) {
	prerelease := true
	tagName := "v1.0.0-beta"

	releases := []*github.RepositoryRelease{
		{
			Prerelease: &prerelease,
			TagName:    &tagName,
		},
	}

	// Without preRelease flag, should return error
	_, err := findLatestRelease(releases, false)
	if err == nil {
		t.Error("findLatestRelease() should return error when only prereleases exist and preRelease=false")
	}

	// With preRelease flag, should return the prerelease
	release, err := findLatestRelease(releases, true)
	if err != nil {
		t.Errorf("findLatestRelease() with preRelease=true should not error: %v", err)
	}
	if release.GetTagName() != tagName {
		t.Errorf("findLatestRelease() returned wrong release: got %s, want %s", release.GetTagName(), tagName)
	}
}

// TestFindLatestRelease_SelectsLatest tests that latest version is selected.
func TestFindLatestRelease_SelectsLatest(t *testing.T) {
	prerelease := false
	tagV1 := "v1.0.0"
	tagV2 := "v2.0.0"
	tagV3 := "v3.0.0"

	releases := []*github.RepositoryRelease{
		{Prerelease: &prerelease, TagName: &tagV1},
		{Prerelease: &prerelease, TagName: &tagV3},
		{Prerelease: &prerelease, TagName: &tagV2},
	}

	release, err := findLatestRelease(releases, false)
	if err != nil {
		t.Fatalf("findLatestRelease() error: %v", err)
	}

	if release.GetTagName() != tagV3 {
		t.Errorf("findLatestRelease() should select latest: got %s, want %s", release.GetTagName(), tagV3)
	}
}

// TestFindAsset_NoAssets tests error when no matching asset found.
func TestFindAsset_NoAssets(t *testing.T) {
	release := &github.RepositoryRelease{
		Assets: []*github.ReleaseAsset{},
	}

	_, err := findAsset(release)
	if err == nil {
		t.Error("findAsset() should return error for empty assets")
	}
}

// TestFindAsset_WrongAssetName tests that wrong asset names are not matched.
func TestFindAsset_WrongAssetName(t *testing.T) {
	wrongName := "wrong-asset.zip"
	release := &github.RepositoryRelease{
		Assets: []*github.ReleaseAsset{
			{Name: &wrongName},
		},
	}

	_, err := findAsset(release)
	if err == nil {
		t.Error("findAsset() should return error when no matching asset")
	}
}

// TestGetAssetName tests the asset name generation.
func TestGetAssetName(t *testing.T) {
	name := getAssetName()

	if name == "" {
		t.Error("getAssetName() should not return empty string")
	}

	// Should contain platform info
	if len(name) < 5 {
		t.Errorf("getAssetName() returned too short name: %s", name)
	}
}
