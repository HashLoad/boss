package models

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"
)

type Dependencyes struct {
	Name    string `ymal:"name"`
	Version string `yaml:"version"`
	Hash    string `yaml:"hash"`
}

type PackageLock struct {
	fileName  string
	Hash      string                    `yaml:"hash"`
	Updated   time.Time                 `yaml:"updated"`
	Installed map[string]string         `yaml:"installedModules"`
	Tree      map[string][]Dependencyes `yaml:"tree"`
}

func LoadPackageLock(parentPackage *Package) PackageLock {

	packageLockPath := filepath.Join(filepath.Dir(parentPackage.fileName), consts.FilePackageLock)
	if fileBytes, err := ioutil.ReadFile(packageLockPath); err != nil {
		hash := md5.New()
		if _, err := io.WriteString(hash, parentPackage.Name); err != nil {
			msg.Warn("Failed on  write machine id to hash")
		}

		return PackageLock{
			fileName: packageLockPath,
			Updated:  time.Now(),
			Hash:     hex.EncodeToString(hash.Sum(nil)),

			Installed: map[string]string{"delphi-docker": "1.3.1"},
			Tree: map[string][]Dependencyes{"delphi-docker": {{
				Name:    "horse",
				Version: "^1.2.8",
				Hash:    "hash-aqui",
			}}},
		}
	} else {
		lockfile := PackageLock{
			fileName: packageLockPath,
		}
		if err := json.Unmarshal(fileBytes, &lockfile); err != nil {
			utils.HandleError(err)
		}
		return lockfile
	}
}

func (p *PackageLock) Save() {
	marshal, err := json.MarshalIndent(&p, "", "\t")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_ = ioutil.WriteFile(p.fileName, marshal, 664)
}
