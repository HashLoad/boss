package core

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"github.com/masterminds/semver"
	"github.com/snakeice/gogress"
)

const latestRelease string = "https://api.github.com/repos/HashLoad/boss/releases/latest"
const releaseTag string = "https://api.github.com/repos/HashLoad/boss/releases/tags/%s"
const tags string = "https://api.github.com/repos/HashLoad/boss/tags"

func DoBossUpgrade(preRelease bool) {
	var link string
	var size float64
	var version string

	if !preRelease {
		link, size, version = getDownloadLink(latestRelease)
	} else {
		tag := getLastTag()
		link, size, version = getDownloadLink(fmt.Sprintf(releaseTag, tag))
	}

	if !checkVersion(version, preRelease) {
		return
	}

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath, _ := filepath.Abs(ex)

	_ = os.Remove(exePath + "_o")

	if err := os.Rename(exePath, exePath+"_o"); err != nil {
		msg.Warn("Failed on rename " + exePath + " to " + exePath + "_0")
	}
	if err := downloadFile(exePath+"_n", link, size); err != nil {
		msg.Err("Failed on download ", err.Error())
		if err := os.Rename(exePath+"_o", exePath); err != nil {
			msg.Err("Failed on rename "+exePath+"_o"+" to "+exePath, err.Error())
		}
	} else {
		if err := os.Rename(exePath+"_n", exePath); err != nil {
			msg.Err("Failed on rename "+exePath+"_n"+" to "+exePath, err.Error())
		}
	}
}

func checkVersion(newVersion string, preRelease bool) bool {
	version, _ := semver.NewVersion(newVersion)
	current, _ := semver.NewVersion(consts.Version)

	needUpdate := version.GreaterThan(current)

	if !needUpdate && preRelease {
		needUpdate = current.Prerelease() == "" && version.Prerelease() != ""
	} else if !needUpdate && !preRelease {
		needUpdate = current.Prerelease() != "" && version.Prerelease() == ""
	}

	if needUpdate {
		println(consts.Version, " -> ", newVersion)
	} else {
		println(newVersion)
		println("already up to date!")
	}
	return needUpdate
}

func getDownloadLink(releaseUrl string) (link string, size float64, version string) {
	resp := makeRequest(releaseUrl)
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg.Die(err.Error())
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(contents, &obj); err != nil {
		msg.Die("failed in parse version JSON")
	}
	for _, value := range obj["assets"].([]interface{}) {
		bossExe := value.(map[string]interface{})
		if bossExe["name"].(string) == "boss.exe" {
			return bossExe["browser_download_url"].(string), bossExe["size"].(float64), obj["tag_name"].(string)
		}
	}
	utils.HandleError(resp.Body.Close())
	msg.Die("not found")
	return "", 0, ""
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
