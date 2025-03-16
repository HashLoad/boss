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

const XMLTagNameProperty string = "PropertyGroup"
const XMLValueAttribute = "value"
const XMLTagNamePropertyAttribute string = "Condition"
const XMLTagNamePropertyAttributeValue string = "'$(Base)'!=''"

const XMLTagNameLibraryPath string = "DCC_UnitSearchPath"

const XMLTagNameCompilerOptions string = "CompilerOptions"
const XMLTagNameSearchPaths string = "SearchPaths"
const XMLTagNameOtherUnitFiles string = "OtherUnitFiles"
const XMLTagNameProjectOptions string = "ProjectOptions"
const XMLTagNameBuildModes string = "BuildModes"
const XMLTagNameItem string = "Item"
const XMLNameAttribute = "Name"

const BossInternalDir = "internal."
const BossInternalDirOld = "{internal}"

const BplIdentifierName = "BplIdentifier.exe"

const RegexArtifacts = "(.*.inc$|.*.pas$|.*.dfm$|.*.fmx$|.*.dcu$|.*.bpl$|.*.dcp$|.*.res$)"

const RegistryBasePath = `Software\Embarcadero\BDS\`

var DefaultPaths = []string{BplFolder, DcuFolder, DcpFolder, BinFolder}
