package librarypath

import (
	"github.com/beevik/etree"
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func updateDprojLibraryPath() {
	var dprojsNames = GetDprojNames()
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
			processCurrentPath(child)
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

func GetDprojNames() []string {
	var result []string
	var matches = 0

	if packageJson, e := models.LoadPackage(false); e != nil {
		panic(e)
	} else {
		if len(packageJson.Projects) > 0 {
			for _, project := range packageJson.Projects {
				result = append(result, env.GetCurrentDir()+string(filepath.Separator)+project)
			}

			result = packageJson.Projects
		} else {
			files, err := ioutil.ReadDir(env.GetCurrentDir())
			if err != nil {
				panic(e)
			}
			for _, file := range files {
				matched, e := regexp.MatchString(".*.dproj$", file.Name())
				if e == nil && matched {
					result = append(result, env.GetCurrentDir()+string(filepath.Separator)+file.Name())
					matches++
				}
			}
		}
	}
	return result
}

func processCurrentPath(node *etree.Element) {
	currentPaths := strings.Split(node.Text(), ";")

	currentPaths = GetNewPaths(currentPaths, false)

	node.SetText(strings.Join(currentPaths, ";"))

}
