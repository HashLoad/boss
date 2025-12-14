// Package upgrade handles ZIP and TAR.GZ archive extraction for Boss updates.
// This file provides utilities for reading files from compressed archives.
package upgrade

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
)

// getAssetFromFile returns the asset from the file
func getAssetFromFile(file *os.File, assetName string) ([]byte, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if strings.HasSuffix(assetName, ".zip") {
		return readFileFromZip(file, assetName, stat)
	}

	return readFileFromTargz(file, assetName)
}

// readFileFromZip reads the file from the zip
func readFileFromZip(file *os.File, assetName string, stat os.FileInfo) ([]byte, error) {
	reader, err := zip.NewReader(file, stat.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to create zip reader: %w", err)
	}

	filePreffix := path.Join(fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH), "boss")

	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, filePreffix) {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open file: %w", err)
			}
			defer rc.Close()

			return io.ReadAll(rc)
		}
	}

	return nil, fmt.Errorf("failed to find asset %s in zip", assetName)
}

// readFileFromTargz reads the file from the tar.gz
func readFileFromTargz(file *os.File, assetName string) ([]byte, error) {
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	filePreffix := path.Join(fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH), "boss")

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar header: %w", err)
		}

		if strings.HasPrefix(header.Name, filePreffix) {
			return io.ReadAll(tarReader)
		}
	}

	return nil, fmt.Errorf("failed to find asset %s in tar.gz", assetName)
}
