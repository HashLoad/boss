package upgrade

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func getAssetFromZip(file *os.File, assetName string) ([]byte, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	reader, err := zip.NewReader(file, stat.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to create zip reader: %w", err)
	}

	filePreffix := path.Join(zipFolder, "boss")

	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, filePreffix) {

			return readZipFile(file)
		}
	}

	return nil, fmt.Errorf("failed to find asset %s in zip", assetName)
}

func readZipFile(zfile *zip.File) ([]byte, error) {
	rc, err := zfile.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer rc.Close()

	return ioutil.ReadAll(rc)
}
