package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
)

func hashByte(contentPtr *[]byte) string {
	contents := *contentPtr
	hasher := md5.New()
	hasher.Write(contents)
	return hex.EncodeToString(hasher.Sum(nil))

}

func HashDir(dir string) string {
	var err error
	var finHash = "b:"
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		fileBytes, _ := ioutil.ReadFile(path)
		fileHash := hashByte(&fileBytes)
		finHash += fileHash
		return nil
	})
	if err != nil {
		os.Exit(1)
	}
	c := []byte(finHash)
	m := hashByte(&c)
	return m
}
