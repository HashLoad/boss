package consts

import "path/filepath"

const FILE_PACKAGE = "boss.json"

const SEPARATOR = string(filepath.Separator)

const FOLDER_DEPENDENCIES = "modules"
const ENV_BOSS_BIN = "." + SEPARATOR + FOLDER_DEPENDENCIES + SEPARATOR + ".bin"

const XML_TAG_NAME_PROPERTY string = "PropertyGroup"
const XML_TAG_NAME_PROPERTY_ATTRIBUTE string = "Condition"
const XML_TAG_NAME_PROPERTY_ATTRIBUTE_VALUE string = "'$(Base)'!=''"

const XML_TAG_NAME_LIBRARY_PATH string = "DCC_UnitSearchPath"

const VERSION string = "v1.6.0"
