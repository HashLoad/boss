package setup

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/installer"
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils"
)

func one() {
	env.GlobalConfiguration().InternalRefreshRate = 5
}

func two() {
	oldPath := filepath.Join(env.GetBossHome(), consts.FolderDependencies, consts.BossInternalDirOld+env.HashDelphiPath())
	newPath := filepath.Join(env.GetBossHome(), consts.FolderDependencies, consts.BossInternalDir+env.HashDelphiPath())
	err := os.Rename(oldPath, newPath)
	if !os.IsNotExist(err) {
		utils.HandleError(err)
	}
}

func three() {
	env.GlobalConfiguration().GitEmbedded = true
	env.GlobalConfiguration().SaveConfiguration()
}

func six() {
	err := os.RemoveAll(env.GetInternalGlobalDir())
	utils.HandleError(err)
}

func seven() {
	bossCfg := filepath.Join(env.GetBossHome(), consts.BossConfigFile)
	if _, err := os.Stat(bossCfg); os.IsNotExist(err) {
		return
	}
	file, err := os.Open(bossCfg)
	utils.HandleError(err)

	data := map[string]any{}

	err = json.NewDecoder(file).Decode(&data)
	utils.HandleError(err)

	auth, found := data["auth"].(map[string]any)
	if !found {
		return
	}

	for key, value := range auth {
		authMap, ok := value.(map[string]interface{})
		if !ok {
			continue
		}

		if user, found := authMap["x"]; found {
			us, err := oldDecrypt(user)
			utils.HandleErrorFatal(err)
			env.GlobalConfiguration().Auth[key].SetUser(us)
		}

		if pass, found := authMap["y"]; found {
			ps, err := oldDecrypt(pass)
			utils.HandleErrorFatal(err)
			env.GlobalConfiguration().Auth[key].SetPass(ps)
		}

		if passPhrase, found := authMap["z"]; found {
			pp, err := oldDecrypt(passPhrase)
			utils.HandleErrorFatal(err)
			env.GlobalConfiguration().Auth[key].SetPassPhrase(pp)
		}
	}
}

func cleanup() {
	env.SetInternal(false)
	env.GlobalConfiguration().LastInternalUpdate = time.Now().AddDate(-1000, 0, 0)
	modulesDir := filepath.Join(env.GetBossHome(), consts.FolderDependencies, env.HashDelphiPath())
	if _, err := os.Stat(modulesDir); os.IsNotExist(err) {
		return
	}

	err := os.Remove(filepath.Join(modulesDir, consts.FilePackageLock))
	utils.HandleError(err)
	modules, err := models.LoadPackage(false)
	if err != nil {
		return
	}

	installer.GlobalInstall([]string{}, modules, false, false)
	env.SetInternal(true)
}

func oldDecrypt(securemess any) (string, error) {
	data, ok := securemess.(string)
	if !ok {
		return "", errors.New("error on convert data to string")
	}

	cipherText, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("error on decode base64: %w", err)
	}

	id, err := machineid.ID()
	if err != nil {
		msg.Err("Error on get machine ID")
		id = "AAAA"
	}

	block, err := aes.NewCipher([]byte(id[:16]))
	if err != nil {
		return "", fmt.Errorf("error on create cipher: %w", err)
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("error on check block size: cipher text block size is too short")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	//nolint:staticcheck // Just use the old decrypt method to migrate the data
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}
