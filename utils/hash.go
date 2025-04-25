package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashload/boss/pkg/msg"
)

func hashByte(contentPtr []byte) string {
	hasher := sha256.New()
	hasher.Write(contentPtr)
	return hex.EncodeToString(hasher.Sum(nil))
}

func HashDir(dir string) string {
	var err error
	var finalHash = ""
	err = filepath.Walk(dir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil && !os.IsNotExist(err) {
			msg.Warn("Failed to read file %s", path)
			return nil
		}

		if os.IsNotExist(err) || fileInfo.IsDir() {
			return nil
		}

		fileBytes, err := os.ReadFile(path)
		if err != nil {
			msg.Warn("Failed to read file %s", path)
			//nolint:nilerr // We don't want to stop the process
			return nil
		}

		fileHash := hashByte(fileBytes)
		finalHash += fileHash
		return nil
	})
	if err != nil {
		msg.Err("Failed to read file %s", dir)
		msg.Err("Error: %s", err)
		msg.Fatal("I can't continue, sorry :(")
	}
	finalHashDigest := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(finalHash)))
	return finalHashDigest
}
