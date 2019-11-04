package consts

import "path/filepath"

const FilePackage = "boss.json"
const FilePackageLock = "boss-lock.json"
const FilePackageLockOld = "boss.lock"

const FolderDependencies = "modules"
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

const Version string = "v3.0.1"

const BossInternalDir = "internal."
const BossInternalDirOld = "{internal}"

const BplIdentifierName = "BplIdentifier.exe"

const RegexArtifacts = "(.*.inc$|.*.pas$|.*.dfm$|.*.fmx$|.*.dcu$|.*.bpl$|.*.dcp$)"

const RegistyBasePath = `Software\Embarcadero\BDS\`

var DefaultPaths = []string{BplFolder, DcuFolder, DcpFolder, BinFolder}
