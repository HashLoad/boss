package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/msg"
	"github.com/spf13/cobra"
	"gopkg.in/cheggaaa/pb.v2"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade a cli",
	Long:  `upgrade a cli`,
	Run: func(cmd *cobra.Command, args []string) {
		err, link, size, version := getLink()
		if err != nil {
			err.Error()
		}

		if !checkVersion(version) {
			return
		}

		exePath, err := filepath.Abs(os.Args[0])
		if err != nil {
			err.Error()
		}
		if err := os.Remove(exePath + "_o"); err != nil {
			msg.Err("Failed to remove old file "+exePath+"_0", err)
		}
		if err := os.Rename(exePath, exePath+"_o"); err != nil {
			msg.Err("Failed on rename "+exePath+" to "+exePath+"_0", err)
		}
		if err := downloadFile(exePath+"_n", link, size); err != nil {
			if err := os.Rename(exePath+"_o", exePath); err != nil {
				msg.Err("Failed on rename "+exePath+"_o"+" to "+exePath, err)
			}
			err.Error()
		} else {
			if err := os.Rename(exePath+"_n", exePath); err != nil {
				msg.Err("Failed on rename "+exePath+"_n"+" to "+exePath, err)
			}
		}
	},
}

func checkVersion(newVersion string) bool {
	version, _ := semver.NewVersion(newVersion)
	current, _ := semver.NewVersion(consts.VERSION)
	needUpdate := version.GreaterThan(current)
	if needUpdate {
		println(consts.VERSION, " -> ", newVersion)
	} else {
		println(consts.VERSION, " <= ", newVersion)
		println("already up to date!")
	}
	return needUpdate
}

func getLink() (err error, link string, size float64, version string) {
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
	bossExe := obj["assets"].([]interface{})[0].(map[string]interface{})
	return nil, bossExe["browser_download_url"].(string), bossExe["size"].(float64), obj["tag_name"].(string)
}

func closeFile(out *os.File) {
	_ = out.Close()
}

func downloadFile(filepath string, url string, size float64) (err error) {
	_ = os.Remove(filepath)
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer closeFile(out)

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

func init() {
	RootCmd.AddCommand(upgradeCmd)
}
