package consts

import "path/filepath"

const FilePackage = "boss.json"

const Separator = string(filepath.Separator)

const FolderDependencies = "modules"
const BinFolder string = ".bin"

const MinimalDependencyVersion string = ">0.0.0"

const EnvBossBin = "." + Separator + FolderDependencies + Separator + BinFolder

const XmlTagNameProperty string = "PropertyGroup"
const XmlTagNamePropertyAttribute string = "Condition"
const XmlTagNamePropertyAttributeValue string = "'$(Base)'!=''"

const XmlTagNameLibraryPath string = "DCC_UnitSearchPath"

const Version string = "v2.0.3"
