package consts

import "path/filepath"

const (
	FilePackage        = "boss.json"
	FilePackageLock    = "boss-lock.json"
	FileBplOrder       = "bpl_order.txt"
	FileExtensionBpl   = ".bpl"
	FileExtensionDcp   = ".dcp"
	FileExtensionDpk   = ".dpk"
	FileExtensionDpr   = ".dpr"
	FileExtensionDproj = ".dproj"
	FileExtensionLpi   = ".lpi"

	FilePackageLockOld = "boss.lock"
	FolderDependencies = "modules"

	FolderEnv = "env"

	FolderEnvBpl = FolderEnv + string(filepath.Separator) + "bpl"
	FolderEnvDcp = FolderEnv + string(filepath.Separator) + "dcp"
	FolderEnvDcu = FolderEnv + string(filepath.Separator) + "dcu"

	FolderBossHome = ".boss"

	BinFolder string = ".bin"
	BplFolder string = ".bpl"
	DcpFolder string = ".dcp"
	DcuFolder string = ".dcu"

	BossConfigFile = "boss.cfg.json"

	MinimalDependencyVersion string = ">0.0.0"

	EnvBossBin = "." + string(filepath.Separator) + FolderDependencies + string(filepath.Separator) + BinFolder

	XMLTagNameProperty               string = "PropertyGroup"
	XMLValueAttribute                       = "value"
	XMLTagNamePropertyAttribute      string = "Condition"
	XMLTagNamePropertyAttributeValue string = "'$(Base)'!=''"

	XMLTagNameLibraryPath string = "DCC_UnitSearchPath"

	XMLTagNameCompilerOptions string = "CompilerOptions"
	XMLTagNameSearchPaths     string = "SearchPaths"
	XMLTagNameOtherUnitFiles  string = "OtherUnitFiles"
	XMLTagNameProjectOptions  string = "ProjectOptions"
	XMLTagNameBuildModes      string = "BuildModes"
	XMLTagNameItem            string = "Item"
	XMLNameAttribute                 = "Name"

	BossInternalDir    = "internal."
	BossInternalDirOld = "{internal}"

	BplIdentifierName = "BplIdentifier.exe"

	RegexArtifacts = "(.*.inc$|.*.pas$|.*.dfm$|.*.fmx$|.*.dcu$|.*.bpl$|.*.dcp$|.*.res$)"

	RegistryBasePath = `Software\Embarcadero\BDS\`

	StatusMsgUpToDate        = "up to date"
	StatusMsgResolvingVer    = "resolving version"
	StatusMsgNoProjects      = "no projects"
	StatusMsgNoBossJSON      = "no boss.json"
	StatusMsgBuildError      = "build error"
	StatusMsgAlreadyUpToDate = "boss is already up to date"

	GitBranchMain   = "main"
	GitBranchMaster = "master"

	GitProtocolSSH = "ssh"
)

type Platform string

const (
	PlatformWin32        Platform = "Win32"
	PlatformWin64        Platform = "Win64"
	PlatformOSX32        Platform = "OSX32"
	PlatformOSX64        Platform = "OSX64"
	PlatformOSXArm64     Platform = "OSXARM64"
	PlatformLinux64      Platform = "Linux64"
	PlatformAndroid      Platform = "Android"
	PlatformAndroid64    Platform = "Android64"
	PlatformiOSDevice32  Platform = "iOSDevice32"
	PlatformiOSDevice64  Platform = "iOSDevice64"
	PlatformiOSSimulator Platform = "iOSSimulator"
	PlatformiOSSimARM64  Platform = "iOSSimARM64"
)

func (p Platform) String() string {
	return string(p)
}

func (p Platform) IsValid() bool {
	switch p {
	case PlatformWin32, PlatformWin64, PlatformOSX32, PlatformOSX64, PlatformOSXArm64,
		PlatformLinux64, PlatformAndroid, PlatformAndroid64,
		PlatformiOSDevice32, PlatformiOSDevice64, PlatformiOSSimulator, PlatformiOSSimARM64:
		return true
	}
	return false
}

func DefaultPlatform() Platform {
	return PlatformWin32
}

func DefaultPaths() []string {
	return []string{BplFolder, DcuFolder, DcpFolder, BinFolder}
}
