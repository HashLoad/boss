package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
)

const (
	SHA1   = "sha1"
	SHA256 = "sha256"
	MD5    = "md5"
)

func getHash(name *string) (hash.Hash, error) {
	if *name == SHA1 {
		return sha1.New(), nil
	} else if *name == SHA256 {
		return sha256.New(), nil
	} else if *name == MD5 {
		return md5.New(), nil
	}
	message := "Hash Algorithm is not supported"
	err := errors.New(message)
	return nil, err

}

func CreateHash(path string, hashAlgorithm string) (string, error) {
	hash, err := getHash(&hashAlgorithm)

	if err != nil {
		return "", nil
	}

	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		_, err = io.WriteString(hash, path)
		HandleError(err)
		return nil
	})

	if err != nil {
		return "", nil
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
