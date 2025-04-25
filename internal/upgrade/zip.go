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

	"github.com/google/go-github/v69/github"
)

func getAssetFromFile(file *os.File, asset *github.ReleaseAsset) ([]byte, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	formatHandler := map[string]func(*os.File, os.FileInfo) ([]byte, error){
		".zip":    readFileFromZip,
		".tar.gz": readFileFromTargz,
	}

	ext := path.Ext(asset.GetName())

	handler, ok := formatHandler[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	return handler(file, stat)
}

func readFileFromZip(file *os.File, stat os.FileInfo) ([]byte, error) {
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

	return nil, fmt.Errorf("failed to find asset %s in zip", file.Name())
}

func readFileFromTargz(file *os.File, _ os.FileInfo) ([]byte, error) {
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

	return nil, fmt.Errorf("failed to find asset %s in tar.gz", file.Name())
}
