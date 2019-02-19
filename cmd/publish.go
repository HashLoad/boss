package cmd

import (
	"archive/tar"
	"bytes"
	"github.com/hashload/boss/core/git"
	"github.com/hashload/boss/models"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4/plumbing/format/gitignore"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var publisCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish package to registry",
	Long:  `Publish package to registry`,
	Run: func(cmd *cobra.Command, args []string) {
		if true {
			log.Println("TODO: complete this feature")
			return
		}

		bossFile, e := models.LoadPackage(false)
		if e != nil {
			e.Error()
		}

		var buff bytes.Buffer
		tw := tar.NewWriter(&buff)

		patternsString := strings.Split(git.Gitignore, "\n")
		patterns := []gitignore.Pattern{}

		for _, value := range patternsString {
			patterns = append(patterns, gitignore.ParsePattern(value, nil))
		}
		_ = filepath.Walk(path.Join("./", bossFile.MainSrc),
			func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}

				for _, pattern := range patterns {
					if pattern.Match([]string{filepath.Base(path)}, info.IsDir()) == gitignore.Exclude {
						return nil
					}
				}

				//git.GetRepository()

				hdr := &tar.Header{
					Name: info.Name(),
					Mode: 0600,
					Size: info.Size(),
				}

				if err := tw.WriteHeader(hdr); err != nil {
					log.Fatal(err)
				}

				if file, err := ioutil.ReadFile(path); err != nil {
					log.Fatal(err)
				} else {
					if _, errWrite := tw.Write(file); errWrite != nil {
						log.Fatal(errWrite)
					}
				}
				return nil
			})

		if err := ioutil.WriteFile("tmp_.tar", buff.Bytes(), os.ModePerm); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(publisCmd)
}
