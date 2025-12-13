package librarypath

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/beevik/etree"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
)

var (
	reProjectFile = regexp.MustCompile(`.*` + regexp.QuoteMeta(consts.FileExtensionDproj) + `|.*` + regexp.QuoteMeta(consts.FileExtensionLpi) + `$`)
	reLazarusFile = regexp.MustCompile(`.*` + regexp.QuoteMeta(consts.FileExtensionLpi) + `$`)
)

// updateDprojLibraryPath updates the library path in the project file
func updateDprojLibraryPath(pkg *domain.Package) {
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

// updateOtherUnitFilesProject updates the other unit files in the project file
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

// processCompilerOptions processes the compiler options
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

// createTagOtherUnitFiles creates the other unit files tag
func createTagOtherUnitFiles(node *etree.Element) *etree.Element {
	child := node.CreateElement(consts.XMLTagNameOtherUnitFiles)
	child.CreateAttr("Value", "")
	return child
}

// updateGlobalBrowsingPath updates the global browsing path
func updateGlobalBrowsingPath(pkg *domain.Package) {
	var isLazarus = isLazarus()
	var projectNames = GetProjectNames(pkg)
	for i, projectName := range projectNames {
		if !isLazarus {
			updateGlobalBrowsingByProject(projectName, i == 0)
		}
	}
}

// updateLibraryPathProject updates the library path in the project file
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

// createTagLibraryPath creates the library path tag
func createTagLibraryPath(node *etree.Element) *etree.Element {
	child := node.CreateElement(consts.XMLTagNameLibraryPath)
	return child
}

// GetProjectNames returns the project names
func GetProjectNames(pkg *domain.Package) []string {
	var result []string

	if len(pkg.Projects) > 0 {
		result = pkg.Projects
	} else {
		files, err := os.ReadDir(env.GetCurrentDir())
		if err != nil {
			panic(err)
		}

		for _, file := range files {
			if reProjectFile.MatchString(file.Name()) {
				result = append(result, filepath.Join(env.GetCurrentDir(), file.Name()))
			}
		}
	}

	return result
}

// isLazarus checks if the project is a Lazarus project
func isLazarus() bool {
	files, err := os.ReadDir(env.GetCurrentDir())
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		matched := reLazarusFile.MatchString(file.Name())
		if matched {
			return true
		}
	}
	return false
}

// processCurrentPath processes the current path
func processCurrentPath(node *etree.Element, rootPath string) {
	currentPaths := strings.Split(node.Text(), ";")

	currentPaths = GetNewPaths(currentPaths, false, rootPath)

	node.SetText(strings.Join(currentPaths, ";"))
}
