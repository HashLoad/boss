package consts

import "path/filepath"

const FilePackage = "boss.json"
const FilePackageLock = "boss-lock.json"
const FileBplOrder = "bpl_order.txt"
const FileExtensionBpl = ".bpl"
const FileExtensionDcp = ".dcp"
const FileExtensionDpk = ".dpk"

const FilePackageLockOld = "boss.lock"
const FolderDependencies = "modules"

const FolderEnv = "env"

var FolderEnvBpl = filepath.Join(FolderEnv, "bpl")
var FolderEnvDcp = filepath.Join(FolderEnv, "dcp")
var FolderEnvDcu = filepath.Join(FolderEnv, "dcu")

const FolderBossHome = ".boss"

const BinFolder string = ".bin"
const BplFolder string = ".bpl"
const DcpFolder string = ".dcp"
const DcuFolder string = ".dcu"

const BossConfigFile = "boss.cfg.json"

const MinimalDependencyVersion string = ">0.0.0"

const EnvBossBin = "." + string(filepath.Separator) + FolderDependencies + string(filepath.Separator) + BinFolder

const XmlTagNameProperty string = "PropertyGroup"
const XmlValueAttribute = "value"
const XmlTagNamePropertyAttribute string = "Condition"
const XmlTagNamePropertyAttributeValue string = "'$(Base)'!=''"

const XmlTagNameLibraryPath string = "DCC_UnitSearchPath"

const XmlTagNameCompilerOptions string = "CompilerOptions"
const XmlTagNameSearchPaths string = "SearchPaths"
const XmlTagNameOtherUnitFiles string = "OtherUnitFiles"

const Version string = "v3.0.5"

const BossInternalDir = "internal."
const BossInternalDirOld = "{internal}"

const BplIdentifierName = "BplIdentifier.exe"

const RegexArtifacts = "(.*.inc$|.*.pas$|.*.dfm$|.*.fmx$|.*.dcu$|.*.bpl$|.*.dcp$)"

const RegistryBasePath = `Software\Embarcadero\BDS\`

var DefaultPaths = []string{BplFolder, DcuFolder, DcpFolder, BinFolder}
