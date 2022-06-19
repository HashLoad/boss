package core

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashload/boss/internal/version"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"github.com/masterminds/semver"
	"github.com/minio/selfupdate"
	"github.com/snakeice/gogress"
)

const latestRelease string = "https://api.github.com/repos/hashload/boss/releases/latest"
const releaseTag string = "https://api.github.com/repos/hashload/boss/releases/tags/%s"
const tags string = "https://api.github.com/repos/hashload/boss/tags"

func DoBossUpgrade(preRelease bool) {
	var link string
	var size float64
	var version string
	var fileName string

	if !preRelease {
		link, size, version, fileName = getDownloadLink(latestRelease)
	} else {
		tag := getLastTag()
		link, size, version, fileName = getDownloadLink(fmt.Sprintf(releaseTag, tag))
	}

	if !checkVersion(version, preRelease) {
		return
	}

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath, _ := filepath.Abs(ex)

	_ = os.Remove(exePath + "_bkp")

	downloadPath := filepath.Join(os.TempDir(), fileName)

	if err := downloadFile(downloadPath, link, size); err != nil {
		msg.Err("Failed on download ", err.Error())
		return
	} else {
		defer os.Remove(downloadPath)
	}

	zipReader, err := zip.OpenReader(downloadPath)
	if err != nil {
		msg.Err("Failed on open zip ", err.Error())
		return
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		// get file name from path
		tmpFilename := filepath.Base(f.Name)

		if strings.HasPrefix(tmpFilename, "boss") {
			rc, err := f.Open()
			if err != nil {
				msg.Err("Failed on open zip file ", err.Error())
				return
			}
			defer rc.Close()

			err = selfupdate.Apply(rc, selfupdate.Options{
				OldSavePath: fmt.Sprintf("%s_bkp", exePath),
				TargetPath:  exePath,
			})
			if err != nil {
				msg.Err("Failed on apply update ", err.Error())
				return
			}

			break
		}
	}
	msg.Info("Upgrade to version " + version + " success")
}

func checkVersion(newVersion string, preRelease bool) bool {
	new, _ := semver.NewVersion(newVersion)
	current, _ := semver.NewVersion(version.Get().Version)

	needUpdate := new.GreaterThan(current)

	if !needUpdate && preRelease {
		needUpdate = current.Prerelease() == "" && new.Prerelease() != ""
	} else if !needUpdate && !preRelease {
		needUpdate = current.Prerelease() != "" && new.Prerelease() == ""
	}

	if needUpdate {
		println(version.Get().Version, " -> ", newVersion)
	} else {
		println(newVersion)
		println("already up to date!")
	}
	return needUpdate
}

func getDownloadLink(releaseUrl string) (string, float64, string, string) {
	resp := makeRequest(releaseUrl)
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg.Die(err.Error())
	}

	fileName := "boss-" + runtime.GOOS + "-" + runtime.GOARCH + ".zip"

	var obj map[string]interface{}
	if err := json.Unmarshal(contents, &obj); err != nil {
		msg.Die("failed in parse version JSON")
	}
	for _, value := range obj["assets"].([]interface{}) {
		bossExe := value.(map[string]interface{})
		if bossExe["name"].(string) == fileName {
			return bossExe["browser_download_url"].(string), bossExe["size"].(float64), obj["tag_name"].(string), fileName
		}
	}
	utils.HandleError(resp.Body.Close())
	msg.Die("not found " + fileName + " in release " + obj["tag_name"].(string))
	return "", 0, "", ""
}

func getLastTag() string {
	resp := makeRequest(tags)
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg.Die(err.Error())
	}
	var obj []interface{}
	if err := json.Unmarshal(contents, &obj); err != nil {
		msg.Die("failed in parse tags JSON")
	}

	utils.HandleError(resp.Body.Close())
	tagObj := obj[0].(map[string]interface{})
	return tagObj["name"].(string)
}

func makeRequest(url string) *http.Response {
	resp, err := http.Get(url)
	if err != nil {
		msg.Die(err.Error())
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		msg.Die("bad status: %s", resp.Status)
	}
	return resp
}

func downloadFile(filepath string, url string, size float64) (err error) {
	_ = os.Remove(filepath)
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	bar := gogress.New64(int64(math.Round(size)))
	bar.Start()
	proxyReader := bar.NewProxyReader(resp.Body)
	defer proxyReader.Close()
	_, err = io.Copy(out, proxyReader)

	utils.HandleError(out.Close())
	utils.HandleError(resp.Body.Close())

	return err
}
