package librarypath

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/hashload/boss/consts"
	bossRegistry "github.com/hashload/boss/core/registry"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"golang.org/x/sys/windows/registry"
)

const BrowsingPathRegistry = "Browsing Path"

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
	processCompilerOptions(compilerOptions)

	projectOptions := root.SelectElement(consts.XmlTagNameProjectOptions)

	buildModes := projectOptions.SelectElement(consts.XmlTagNameBuildModes)
	for _, item := range buildModes.SelectElements(consts.XmlTagNameItem) {
		attribute := item.SelectAttr(consts.XmlNameAttribute)
		compilerOptions := item.SelectElement(consts.XmlTagNameCompilerOptions)
		if compilerOptions != nil {
			msg.Info("  Updating %s mode", attribute.Value)
			processCompilerOptions(compilerOptions)
		}
	}

	doc.WriteSettings.CanonicalAttrVal = true
	doc.WriteSettings.CanonicalEndTags = false
	doc.WriteSettings.CanonicalText = true
	if err := doc.WriteToFile(lpiName); err != nil {
		panic(err)
	}
}

func processCompilerOptions(compilerOptions *etree.Element) {
	searchPaths := compilerOptions.SelectElement(consts.XmlTagNameSearchPaths)
	if searchPaths == nil {
		return
	}
	otherUnitFiles := searchPaths.SelectElement(consts.XmlTagNameOtherUnitFiles)
	if otherUnitFiles == nil {
		otherUnitFiles = createTagOtherUnitFiles(searchPaths)
	}
	value := otherUnitFiles.SelectAttr("Value")
	currentPaths := strings.Split(value.Value, ";")
	currentPaths = GetNewPaths(currentPaths, false, env.GetCurrentDir())
	value.Value = strings.Join(currentPaths, ";")
}

func createTagOtherUnitFiles(node *etree.Element) *etree.Element {
	child := node.CreateElement(consts.XmlTagNameOtherUnitFiles)
	child.CreateAttr("Value", "")
	return child
}

func updateGlobalBrowsingByProject(dprojName string) {
	ideVersion := bossRegistry.GetCurrentDelphiVersion()
	if ideVersion == "" {
		msg.Err("Version not found for path %s", env.GlobalConfiguration.DelphiPath)
	}
	library, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath+ideVersion+`\Library`, registry.ALL_ACCESS)

	if err != nil {
		msg.Err(`Registry path` + consts.RegistryBasePath + ideVersion + `\Library not exists`)
		return
	}

	libraryInfo, err := library.Stat()
	if err != nil {
		msg.Err(err.Error())
		return
	}
	platforms, err := library.ReadSubKeyNames(int(libraryInfo.SubKeyCount))
	if err != nil {
		msg.Err("No platform found for delphi " + ideVersion)
		return
	}

	for _, platform := range platforms {
		delphiPlatform, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath+ideVersion+`\Library\`+platform, registry.ALL_ACCESS)
		utils.HandleError(err)
		paths, _, err := delphiPlatform.GetStringValue(BrowsingPathRegistry)
		if err != nil {
			msg.Debug("Failed to update library path from platform %s with delphi %s", platform, ideVersion)
			continue
		}

		splitPaths := strings.Split(paths, ";")
		rootPath := filepath.Join(env.GetCurrentDir(), path.Dir(dprojName))
		newSplitPaths := GetNewBrowsingPaths(splitPaths, false, rootPath)
		newPaths := strings.Join(newSplitPaths, ";")
		err = delphiPlatform.SetStringValue(BrowsingPathRegistry, newPaths)
		utils.HandleError(err)
	}
}

func updateGlobalBrowsingPath(pkg *models.Package) {
	var isLazarus = isLazarus()
	var projectNames = GetProjectNames(pkg)
	for _, projectName := range projectNames {
		if !isLazarus {
			updateGlobalBrowsingByProject(projectName)
		}
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
				child = createTagLibraryPath(children)
			}
			rootPath := filepath.Join(env.GetCurrentDir(), path.Dir(dprojName))
			if _, err := os.Stat(rootPath); os.IsNotExist(err) {
				rootPath = env.GetCurrentDir()
			}
			processCurrentPath(child, rootPath)
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
		result = pkg.Projects
	} else {
		files, err := ioutil.ReadDir(env.GetCurrentDir())
		if err != nil {
			panic(err)
		}

		regex := regexp.MustCompile(".*.dproj|.*.lpi$")

		for _, file := range files {
			matched := regex.MatchString(file.Name())
			if matched {
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

	r := regexp.MustCompile(".*.lpi$")

	for _, file := range files {
		matched := r.MatchString(file.Name())
		if matched {
			return true
		}
	}
	return false
}

func processCurrentPath(node *etree.Element, rootPath string) {
	currentPaths := strings.Split(node.Text(), ";")

	currentPaths = GetNewPaths(currentPaths, false, rootPath)

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
