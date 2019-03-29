package utils

import (
	"github.com/beevik/etree"
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func UpdateLibraryPath() {
	var dprojsNames = getDprojName()
	for _, dprojName := range dprojsNames {
		updateLibraryPathProject(dprojName)
	}
}

func updateLibraryPathProject(dprojName string) {
	doc := etree.NewDocument()
	info, err := os.Stat(dprojName)
	if os.IsNotExist(err) || info.IsDir() {
		msg.Err(".dproj not found.")
		return
	}
	e := doc.ReadFromFile(dprojName)
	if e != nil {
		msg.Err("Error on read dproj: %s", e)
		return
	}
	root := doc.Root()

	childrens := root.FindElements(consts.XmlTagNameProperty)
	for _, children := range childrens {
		attribute := children.SelectAttr(consts.XmlTagNamePropertyAttribute)
		if attribute != nil && attribute.Value == consts.XmlTagNamePropertyAttributeValue {
			child := children.SelectElement(consts.XmlTagNameLibraryPath)
			if child == nil {
				child = createTag(children)
			}
			processCurrentPathpaths(child)
		}
	}

	doc.WriteSettings.CanonicalAttrVal = true
	doc.WriteSettings.CanonicalEndTags = true
	doc.WriteSettings.CanonicalText = true

	if err := doc.WriteToFile(dprojName); err != nil {
		panic(err)
	}

}

func createTag(node *etree.Element) *etree.Element {
	child := node.CreateElement(consts.XmlTagNameLibraryPath)
	return child
}

func getDprojName() []string {

	var result []string
	var matches = 0
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if packageJson, e := models.LoadPackage(false); e != nil {
		panic(e)
	} else {
		if len(packageJson.Projects) > 0 {
			for _, project := range packageJson.Projects {
				result = append(result, dir+string(filepath.Separator)+project)
			}

			result = packageJson.Projects
		} else {
			files, err := ioutil.ReadDir(dir)
			if err != nil {
				panic(e)
			}
			for _, file := range files {
				matched, e := regexp.MatchString(".*.dproj$", file.Name())
				if e == nil && matched {
					result = append(result, dir+string(filepath.Separator)+file.Name())
					matches++
				}
			}
		}
	}
	return result
}

func removeIndex(array []string, index int) []string {
	return append(array[:index], array[index+1:]...)
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func getNewPaths(paths []string) []string {
	dir, _ := os.Getwd()
	path := filepath.Join(dir, consts.FolderDependencies)
	_, e := os.Stat(path)
	if os.IsNotExist(e) {
		return nil
	}
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		curr, _ := os.Getwd()
		matched, _ := regexp.MatchString(".*.pas$", info.Name())
		if matched {
			dir, _ := filepath.Split(path)
			dir, _ = filepath.Rel(curr, dir)
			if !Contains(paths, dir) {
				paths = append(paths, dir)
			}
		}
		return nil
	})
	return paths
}

func processCurrentPathpaths(node *etree.Element) {
	currentPaths := strings.Split(node.Text(), ";")
	for index, path := range currentPaths {
		if strings.HasPrefix(strings.Trim(path, ""), consts.FolderDependencies) {
			removeIndex(currentPaths, index)
		}
	}

	currentPaths = getNewPaths(currentPaths)

	node.SetText(strings.Join(currentPaths, ";"))

}
