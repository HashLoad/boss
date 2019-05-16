package consts

import "path/filepath"

const FilePackage = "boss.json"
const FilePackageLock = "boss.lock"

const FolderDependencies = "modules"
const FolderBossHome = ".boss"

const BinFolder string = ".bin"
const BplFolder string = ".bpl"
const DcpFolder string = ".dcp"
const DcuFolder string = ".dcu"

const BossConfigFile = "boss.cfg.json"

const MinimalDependencyVersion string = ">0.0.0"

var EnvBossBin = "." + string(filepath.Separator) + FolderDependencies + string(filepath.Separator) + BinFolder

const XmlTagNameProperty string = "PropertyGroup"
const XmlTagNamePropertyAttribute string = "Condition"
const XmlTagNamePropertyAttributeValue string = "'$(Base)'!=''"

const XmlTagNameLibraryPath string = "DCC_UnitSearchPath"

const Version string = "v2.5.0"

const BossInternalDir = "{internal}"

const BplIdentifierName = "BplIdentifier.exe"
