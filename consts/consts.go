package consts

import "path/filepath"

const FILE_PACKAGE = "boss.json"

const SEPARATOR = string(filepath.Separator)

const FOLDER_DEPENDENCIES = "modules"
const BIN_FOLDER string = ".bin"

const MINIMAL_DEPENDENCY_VERSION string = "^0.0.0"

const ENV_BOSS_BIN = "." + SEPARATOR + FOLDER_DEPENDENCIES + SEPARATOR + BIN_FOLDER

const XML_TAG_NAME_PROPERTY string = "PropertyGroup"
const XML_TAG_NAME_PROPERTY_ATTRIBUTE string = "Condition"
const XML_TAG_NAME_PROPERTY_ATTRIBUTE_VALUE string = "'$(Base)'!=''"

const XML_TAG_NAME_LIBRARY_PATH string = "DCC_UnitSearchPath"

const VERSION string = "v2.0.2"
