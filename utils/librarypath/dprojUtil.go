package librarypath

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
)

func updateDprojLibraryPath(pkg *models.Package) {
	var isLazarus = isLazarus()
	var projectNames = GetProjectNames(pkg)
	for _, projectName := range projectNames {
		if isLazarus {
			updateOtherUnitFilesProject(projectName)
		} else {
			updateLibraryPathProject(projectName)
		}
	}
}

func updateOtherUnitFilesProject(lpiName string) {
	doc := etree.NewDocument()
	info, err := os.Stat(lpiName)
	if os.IsNotExist(err) || info.IsDir() {
		msg.Err(".lpi not found.")
		return
	}
	e := doc.ReadFromFile(lpiName)
	if e != nil {
		msg.Err("Error on read lpi: %s", e)
		return
	}
	root := doc.Root()
	compilerOptions := root.SelectElement(consts.XmlTagNameCompilerOptions)
	searchPaths := compilerOptions.SelectElement(consts.XmlTagNameSearchPaths)
	otherUnitFiles := searchPaths.SelectElement(consts.XmlTagNameOtherUnitFiles)
	if otherUnitFiles == nil {
		otherUnitFiles = createTagOtherUnitFiles(searchPaths)
	}
	value := otherUnitFiles.SelectAttr("Value")
	currentPaths := strings.Split(value.Value, ";")
	currentPaths = GetNewPaths(currentPaths, false)
	value.Value = strings.Join(currentPaths, ";")
	doc.WriteSettings.CanonicalAttrVal = true
	doc.WriteSettings.CanonicalEndTags = false
	doc.WriteSettings.CanonicalText = true
	if err := doc.WriteToFile(lpiName); err != nil {
		panic(err)
	}
}

func createTagOtherUnitFiles(node *etree.Element) *etree.Element {
	child := node.CreateElement(consts.XmlTagNameOtherUnitFiles)
	child.CreateAttr("Value", "")
	return child
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
				child = createTagLibraryPath(children)
			}
			processCurrentPath(child)
		}
	}

	doc.WriteSettings.CanonicalAttrVal = true
	doc.WriteSettings.CanonicalEndTags = false
	doc.WriteSettings.CanonicalText = true

	if err := doc.WriteToFile(dprojName); err != nil {
		panic(err)
	}
}

func createTagLibraryPath(node *etree.Element) *etree.Element {
	child := node.CreateElement(consts.XmlTagNameLibraryPath)
	return child
}

func GetProjectNames(pkg *models.Package) []string {
	var result []string
	var matches = 0

	if len(pkg.Projects) > 0 {
		for _, project := range pkg.Projects {
			result = append(result, env.GetCurrentDir()+string(filepath.Separator)+project)
		}

		result = pkg.Projects
	} else {
		files, err := ioutil.ReadDir(env.GetCurrentDir())
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			matched, e := regexp.MatchString(".*.dproj|.*.lpi$", file.Name())
			if e == nil && matched {
				result = append(result, env.GetCurrentDir()+string(filepath.Separator)+file.Name())
				matches++
			}
		}
	}

	return result
}

func isLazarus() bool {
	files, err := ioutil.ReadDir(env.GetCurrentDir())
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		matched, e := regexp.MatchString(".*.lpi$", file.Name())
		if e == nil && matched {
			return true
		}
	}
	return false
}

func processCurrentPath(node *etree.Element) {
	currentPaths := strings.Split(node.Text(), ";")

	currentPaths = GetNewPaths(currentPaths, false)

	node.SetText(strings.Join(currentPaths, ";"))
}

func _(dprojName string) []string {
	doc := etree.NewDocument()
	info, err := os.Stat(dprojName)
	if os.IsNotExist(err) || info.IsDir() {
		msg.Err(".dproj not found.")
		return []string{}
	}
	err = doc.ReadFromFile(dprojName)
	if err != nil {
		msg.Err("Error on read dproj: %s", err)
		return []string{}
	}
	root := doc.Root()

	var result []string

	path, err := etree.CompilePath("/Project/ProjectExtensions/BorlandProject/Platforms")
	utils.HandleError(err)
	platforms := root.FindElementPath(path)
	for _, platform := range platforms.ChildElements() {
		value := platform.SelectAttr(consts.XmlValueAttribute)
		activePlatform, err := strconv.ParseBool(platform.Text())
		utils.HandleError(err)

		if value != nil && activePlatform {
			result = append(result, value.Value)
		}
	}

	return result

}
