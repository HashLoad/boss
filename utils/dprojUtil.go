package utils

import (
	"errors"
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
	var dprojName = getDprojName()
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

	childrens := root.FindElements(consts.XML_TAG_NAME_PROPERTY)
	for _, children := range childrens {
		attribute := children.SelectAttr(consts.XML_TAG_NAME_PROPERTY_ATTRIBUTE)
		if attribute != nil && attribute.Value == consts.XML_TAG_NAME_PROPERTY_ATTRIBUTE_VALUE {
			child := children.SelectElement(consts.XML_TAG_NAME_LIBRARY_PATH)
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
	child := node.CreateElement(consts.XML_TAG_NAME_LIBRARY_PATH)
	//node.AppendChild(child)
	return child
}

func getDprojName() string {

	var result string
	var matches = 0
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if packageJson, e := models.LoadPackage(false); e != nil {
		panic(e)
	} else {
		if packageJson.DprojName != "" {
			result = packageJson.DprojName
		} else {
			files, err := ioutil.ReadDir(dir)
			if err != nil {
				panic(e)
			}
			for _, file := range files {
				matched, e := regexp.MatchString(".*.dproj$", file.Name())
				if e == nil && matched {
					result = file.Name()
					matches++
				}
			}
		}
	}

	if matches > 1 {
		panic(errors.New("ambiguous projects in same folder!"))
	}

	return dir + string(filepath.Separator) + result
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
	path := filepath.Join(dir, consts.FOLDER_DEPENDENCIES)
	_, e := os.Stat(path)
	if os.IsNotExist(e) {
		return nil
	}
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {

		matched, _ := regexp.MatchString(".*.pas$", info.Name())
		if matched {
			dir, _ := filepath.Split(path)
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
		if strings.HasPrefix(strings.Trim(path, ""), consts.FOLDER_DEPENDENCIES) {
			removeIndex(currentPaths, index)
		}
	}

	currentPaths = getNewPaths(currentPaths)

	node.SetText(strings.Join(currentPaths, ";"))

}
