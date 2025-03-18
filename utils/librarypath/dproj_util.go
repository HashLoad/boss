package librarypath

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/beevik/etree"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
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
	err = doc.ReadFromFile(lpiName)
	if err != nil {
		msg.Err("Error on read lpi: %s", err)
		return
	}

	root := doc.Root()

	compilerOptions := root.SelectElement(consts.XMLTagNameCompilerOptions)
	processCompilerOptions(compilerOptions)

	projectOptions := root.SelectElement(consts.XMLTagNameProjectOptions)

	buildModes := projectOptions.SelectElement(consts.XMLTagNameBuildModes)
	for _, item := range buildModes.SelectElements(consts.XMLTagNameItem) {
		attribute := item.SelectAttr(consts.XMLNameAttribute)
		compilerOptions = item.SelectElement(consts.XMLTagNameCompilerOptions)
		if compilerOptions != nil {
			msg.Info("  Updating %s mode", attribute.Value)
			processCompilerOptions(compilerOptions)
		}
	}

	doc.WriteSettings.CanonicalAttrVal = true
	doc.WriteSettings.CanonicalEndTags = false
	doc.WriteSettings.CanonicalText = true

	if err = doc.WriteToFile(lpiName); err != nil {
		panic(err)
	}
}

func processCompilerOptions(compilerOptions *etree.Element) {
	searchPaths := compilerOptions.SelectElement(consts.XMLTagNameSearchPaths)
	if searchPaths == nil {
		return
	}
	otherUnitFiles := searchPaths.SelectElement(consts.XMLTagNameOtherUnitFiles)
	if otherUnitFiles == nil {
		otherUnitFiles = createTagOtherUnitFiles(searchPaths)
	}
	value := otherUnitFiles.SelectAttr("Value")
	currentPaths := strings.Split(value.Value, ";")
	currentPaths = GetNewPaths(currentPaths, false, env.GetCurrentDir())
	value.Value = strings.Join(currentPaths, ";")
}

func createTagOtherUnitFiles(node *etree.Element) *etree.Element {
	child := node.CreateElement(consts.XMLTagNameOtherUnitFiles)
	child.CreateAttr("Value", "")
	return child
}

func updateGlobalBrowsingPath(pkg *models.Package) {
	var isLazarus = isLazarus()
	var projectNames = GetProjectNames(pkg)
	for i, projectName := range projectNames {
		if !isLazarus {
			updateGlobalBrowsingByProject(projectName, i == 0)
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
	err = doc.ReadFromFile(dprojName)
	if err != nil {
		msg.Err("Error on read dproj: %s", err)
		return
	}
	root := doc.Root()

	childrens := root.FindElements(consts.XMLTagNameProperty)
	for _, children := range childrens {
		attribute := children.SelectAttr(consts.XMLTagNamePropertyAttribute)
		if attribute != nil && attribute.Value == consts.XMLTagNamePropertyAttributeValue {
			child := children.SelectElement(consts.XMLTagNameLibraryPath)
			if child == nil {
				child = createTagLibraryPath(children)
			}
			rootPath := filepath.Join(env.GetCurrentDir(), path.Dir(dprojName))
			if _, err = os.Stat(rootPath); os.IsNotExist(err) {
				rootPath = env.GetCurrentDir()
			}
			processCurrentPath(child, rootPath)
		}
	}

	doc.WriteSettings.CanonicalAttrVal = true
	doc.WriteSettings.CanonicalEndTags = false
	doc.WriteSettings.CanonicalText = true

	if err = doc.WriteToFile(dprojName); err != nil {
		panic(err)
	}
}

func createTagLibraryPath(node *etree.Element) *etree.Element {
	child := node.CreateElement(consts.XMLTagNameLibraryPath)
	return child
}

func GetProjectNames(pkg *models.Package) []string {
	var result []string
	var matches = 0

	if len(pkg.Projects) > 0 {
		result = pkg.Projects
	} else {
		files, err := os.ReadDir(env.GetCurrentDir())
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
	files, err := os.ReadDir(env.GetCurrentDir())
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
