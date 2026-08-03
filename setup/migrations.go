package setup

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/hashload/boss/internal/core/services/installer"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/pkg/pkgmanager"
)

// one sets the internal refresh rate to 5.
func one() {
	env.GlobalConfiguration().InternalRefreshRate = 5
}

// two renames the old internal directory to the new one.
func two() {
	oldPath := filepath.Join(env.GetBossHome(), consts.FolderDependencies, consts.BossInternalDirOld+env.HashDelphiPath())
	newPath := filepath.Join(env.GetBossHome(), consts.FolderDependencies, consts.BossInternalDir+env.HashDelphiPath())
	if err := os.Rename(oldPath, newPath); err != nil && !os.IsNotExist(err) {
		msg.Warn("⚠️ Migration 2: could not rename internal directory: %v", err)
	}
}

// three sets the git embedded to true.
func three() {
	env.GlobalConfiguration().GitEmbedded = true
	env.GlobalConfiguration().SaveConfiguration()
}

// six removes the internal global directory.
func six() {
	if err := os.RemoveAll(env.GetInternalGlobalDir()); err != nil {
		msg.Warn("⚠️ Migration 6: could not remove internal global directory: %v", err)
	}
}

// seven migrates the auth configuration from the legacy AES-CFB fields
// (x/y/z, written by Boss up to v3.0.12) to the current ones.
//
// It reads the legacy values off the in-memory configuration rather than
// re-reading the file: by the time migrations run, the configuration may
// already have been saved once during startup, and reading from disk would
// only ever see what the current struct serialises.
//
// A credential that cannot be decrypted is dropped with a warning instead of
// aborting: an undecryptable legacy value is recoverable with `boss login`,
// while dying here leaves the user with no way to run Boss at all.
func seven() {
	configuration := env.GlobalConfiguration()
	migrated := false

	for repo, auth := range configuration.Auth {
		if migrateLegacyAuth(repo, auth) {
			migrated = true
		}
	}

	if migrated {
		configuration.SaveConfiguration()
	}
}

// migrateLegacyAuth converts one entry's legacy credentials and clears them,
// reporting whether anything was converted.
func migrateLegacyAuth(repo string, auth *env.Auth) bool {
	if auth == nil {
		return false
	}
	if auth.LegacyUser == "" && auth.LegacyPass == "" && auth.LegacyPassPhrase == "" {
		return false
	}

	if decrypted, ok := decryptLegacy(repo, "user", auth.LegacyUser); ok {
		auth.SetUser(decrypted)
	}
	if decrypted, ok := decryptLegacy(repo, "password", auth.LegacyPass); ok {
		auth.SetPass(decrypted)
	}
	// An empty passphrase means the key has none, and storing it back would only
	// re-create the value migration 7 exists to retire.
	if decrypted, ok := decryptLegacy(repo, "passphrase", auth.LegacyPassPhrase); ok && decrypted != "" {
		auth.SetPassPhrase(decrypted)
	}

	auth.LegacyUser, auth.LegacyPass, auth.LegacyPassPhrase = "", "", ""

	return true
}

// decryptLegacy decrypts one legacy value, warning instead of aborting when it
// cannot be read. It reports false when there was nothing to convert.
func decryptLegacy(repo, field, value string) (string, bool) {
	if value == "" {
		return "", false
	}

	decrypted, err := oldDecrypt(value)
	if err != nil {
		msg.Warn("⚠️ Migration 7: could not migrate the %s for %s: %v", field, repo, err)
		return "", false
	}

	return decrypted, true
}

// cleanup cleans up the internal global directory.
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
	modules, err := pkgmanager.LoadPackage()
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
