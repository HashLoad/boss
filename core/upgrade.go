package core

import (
	"encoding/json"
	"fmt"
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/msg"
	"github.com/masterminds/semver"
	"gopkg.in/cheggaaa/pb.v2"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
)

func DoBossUpgrade(preRelease bool) {
	var err error
	var link string
	var size float64
	var version string

	if !preRelease {
		err, link, size, version = getLastestLink()
	} else {
		err, link, size, version = getLinkPreRelease()
	}
	if err != nil {
		err.Error()
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

func getLastestLink() (err error, link string, size float64, version string) {
	resp, err := http.Get("https://api.github.com/repos/HashLoad/boss/releases/latest")
	if err != nil {
		return err, "", 0, ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status), "", 0, ""
	}
	contents, err := ioutil.ReadAll(resp.Body)
	var obj map[string]interface{}
	if err := json.Unmarshal(contents, &obj); err != nil {
		msg.Die("failed in parse version JSON")
	}
	for _, value := range obj["assets"].([]interface{}) {
		bossExe := value.(map[string]interface{})
		if bossExe["name"].(string) == "boss.exe" {
			return nil, bossExe["browser_download_url"].(string), bossExe["size"].(float64), obj["tag_name"].(string)
		}
	}
	return fmt.Errorf("not found"), "", 0, ""
}

func getLinkPreRelease() (err error, link string, size float64, version string) {
	err, tag := getLastTag()
	if err != nil {
		return err, "", 0, ""
	}

	resp, err := http.Get("https://api.github.com/repos/HashLoad/boss/releases/tags/" + tag)
	if err != nil {
		return err, "", 0, ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status), "", 0, ""
	}
	contents, err := ioutil.ReadAll(resp.Body)
	var obj map[string]interface{}
	if err := json.Unmarshal(contents, &obj); err != nil {
		msg.Die("failed in parse version JSON")
	}
	for _, value := range obj["assets"].([]interface{}) {
		bossExe := value.(map[string]interface{})
		if bossExe["name"].(string) == "boss.exe" {
			return nil, bossExe["browser_download_url"].(string), bossExe["size"].(float64), obj["tag_name"].(string)
		}
	}
	return fmt.Errorf("not found"), "", 0, ""
}

func getLastTag() (err error, tag string) {
	resp, err := http.Get("https://api.github.com/repos/HashLoad/boss/tags")
	if err != nil {
		return err, ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status), ""
	}
	contents, err := ioutil.ReadAll(resp.Body)
	var obj []interface{}
	if err := json.Unmarshal(contents, &obj); err != nil {
		msg.Die("failed in parse version JSON")
	}

	tagObj := obj[0].(map[string]interface{})
	//for _, value := range obj["assets"].([]interface{}) {
	//	bossExe := value.(map[string]interface{})
	//	if bossExe["name"].(string) == "boss.exe" {
	//		return nil, bossExe["browser_download_url"].(string), bossExe["size"].(float64), obj["tag_name"].(string)
	//	}
	//}
	return nil, tagObj["name"].(string)
}

//noinspection GoUnhandledErrorResult
func downloadFile(filepath string, url string, size float64) (err error) {
	_ = os.Remove(filepath)
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	bar := pb.New(int(math.Round(size)))
	bar.Start()
	proxyReader := bar.NewProxyReader(resp.Body)
	_, err = io.Copy(out, proxyReader)
	bar.Finish()
	if err != nil {
		return err
	}

	return nil
}
