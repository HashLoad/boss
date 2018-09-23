package utils

import (
	"encoding/xml"
	"os"
)

func LoadPackage(file string) {
	xmlFile, _ := os.Open("Project1.dproj")
	decoder := xml.NewDecoder(xmlFile)
	token, _ := decoder.RawToken()
	print(token)

}
