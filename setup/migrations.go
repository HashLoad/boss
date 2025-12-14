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
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/installer"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
)

// one sets the internal refresh rate to 5
func one() {
	env.GlobalConfiguration().InternalRefreshRate = 5
}

// two renames the old internal directory to the new one
func two() {
	oldPath := filepath.Join(env.GetBossHome(), consts.FolderDependencies, consts.BossInternalDirOld+env.HashDelphiPath())
	newPath := filepath.Join(env.GetBossHome(), consts.FolderDependencies, consts.BossInternalDir+env.HashDelphiPath())
	if err := os.Rename(oldPath, newPath); err != nil && !os.IsNotExist(err) {
		msg.Warn("⚠️ Migration 2: could not rename internal directory: %v", err)
	}
}

// three sets the git embedded to true
func three() {
	env.GlobalConfiguration().GitEmbedded = true
	env.GlobalConfiguration().SaveConfiguration()
}

// six removes the internal global directory
func six() {
	if err := os.RemoveAll(env.GetInternalGlobalDir()); err != nil {
		msg.Warn("⚠️ Migration 6: could not remove internal global directory: %v", err)
	}
}

// seven migrates the auth configuration
func seven() {
	bossCfg := filepath.Join(env.GetBossHome(), consts.BossConfigFile)
	if _, err := os.Stat(bossCfg); os.IsNotExist(err) {
		return
	}
	file, err := os.Open(bossCfg)
	if err != nil {
		msg.Warn("⚠️ Migration 7: could not open config file: %v", err)
		return
	}
	defer file.Close()

	data := map[string]any{}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		msg.Warn("⚠️ Migration 7: could not decode config: %v", err)
		return
	}

	auth, found := data["auth"].(map[string]any)
	if !found {
		return
	}

	for key, value := range auth {
		authMap, ok := value.(map[string]any)
		if !ok {
			continue
		}

		if user, found := authMap["x"]; found {
			decryptedUser, err := oldDecrypt(user)
			if err != nil {
				msg.Die("❌ Migration 7: critical - failed to decrypt user for %s: %v", key, err)
			}
			env.GlobalConfiguration().Auth[key].SetUser(decryptedUser)
		}

		if pass, found := authMap["y"]; found {
			decryptedPassword, err := oldDecrypt(pass)
			if err != nil {
				msg.Die("❌ Migration 7: critical - failed to decrypt password for %s: %v", key, err)
			}
			env.GlobalConfiguration().Auth[key].SetPass(decryptedPassword)
		}

		if passPhrase, found := authMap["z"]; found {
			decryptedPassPhrase, err := oldDecrypt(passPhrase)
			if err != nil {
				msg.Die("❌ Migration 7: critical - failed to decrypt passphrase for %s: %v", key, err)
			}
			env.GlobalConfiguration().Auth[key].SetPassPhrase(decryptedPassPhrase)
		}
	}
}

// cleanup cleans up the internal global directory
func cleanup() {
	env.SetInternal(false)
	env.GlobalConfiguration().LastInternalUpdate = time.Now().AddDate(-1000, 0, 0)
	modulesDir := filepath.Join(env.GetBossHome(), consts.FolderDependencies, env.HashDelphiPath())
	if _, err := os.Stat(modulesDir); os.IsNotExist(err) {
		return
	}

	if err := os.Remove(filepath.Join(modulesDir, consts.FilePackageLock)); err != nil && !os.IsNotExist(err) {
		msg.Debug("Cleanup: could not remove lock file: %v", err)
	}
	modules, err := domain.LoadPackage(false)
	if err != nil {
		return
	}

	installer.GlobalInstall(env.GlobalConfiguration(), []string{}, modules, false, false)
	env.SetInternal(true)
}

// oldDecrypt decrypts the data using the old method for migration purposes.
// This is only used during migration 7 to convert old encrypted credentials.
func oldDecrypt(secureMessage any) (string, error) {
	data, ok := secureMessage.(string)
	if !ok {
		return "", errors.New("error on convert data to string")
	}

	cipherText, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("error on decode base64: %w", err)
	}

	id, err := machineid.ID()
	if err != nil {
		msg.Err("❌ Error on get machine ID")
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

	//nolint:staticcheck,deprecation // Just use the old decrypt method to migrate the data
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}
