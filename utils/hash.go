package utils

import (
	//nolint:gosec // MD5 is used for hash comparison
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/hashload/boss/pkg/msg"
)

func hashByte(contentPtr *[]byte) string {
	contents := *contentPtr
	//nolint:gosec // MD5 is used for hash comparison
	hasher := md5.New()
	hasher.Write(contents)
	return hex.EncodeToString(hasher.Sum(nil))
}

func HashDir(dir string) string {
	var err error
	var finalHash = "b:"
	err = filepath.Walk(dir, func(path string, _ os.FileInfo, err error) error {
		if err != nil && !os.IsNotExist(err) {
			msg.Warn("Failed to read file %s", path)
			return nil
		}

		if os.IsNotExist(err) {
			return nil
		}

		fileBytes, _ := os.ReadFile(path)
		fileHash := hashByte(&fileBytes)
		finalHash += fileHash
		return nil
	})
	if err != nil {
		os.Exit(1)
	}
	c := []byte(finalHash)
	m := hashByte(&c)
	return m
}
