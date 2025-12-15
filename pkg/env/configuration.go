package env

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	sshGit "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils/crypto"
	"golang.org/x/crypto/ssh"
)

// Configuration represents the global configuration for Boss.
// This struct implements the ConfigProvider interface for dependency injection.
// See pkg/env/interfaces.go for interface details.
//
// The configuration is loaded once at startup and injected throughout
// the application via the ConfigProvider interface.
type Configuration struct {
	path                string           `json:"-"`
	Key                 string           `json:"id"`
	Auth                map[string]*Auth `json:"auth"`
	PurgeTime           int              `json:"purge_after"`
	InternalRefreshRate int              `json:"internal_refresh_rate"`
	LastPurge           time.Time        `json:"last_purge_cache"`
	LastInternalUpdate  time.Time        `json:"last_internal_update"`
	DelphiPath          string           `json:"delphi_path,omitempty"`
	ConfigVersion       int64            `json:"config_version"`
	GitEmbedded         bool             `json:"git_embedded"`
	GitShallow          bool             `json:"git_shallow,omitempty"`

	Advices struct {
		SetupPath bool `json:"setup_path,omitempty"`
	} `json:"advices"`
}

// Auth represents authentication credentials for a repository.
type Auth struct {
	UseSSH     bool   `json:"use,omitempty"`
	Path       string `json:"path,omitempty"`
	User       string `json:"user,omitempty"`
	Pass       string `json:"pass,omitempty"`
	PassPhrase string `json:"keypass,omitempty"`
}

// GetUser returns the decrypted username.
func (a *Auth) GetUser() string {
	ret, err := crypto.Decrypt(crypto.MachineKey(), a.User)
	if err != nil {
		msg.Err("❌ Failed to decrypt user.")
		return ""
	}
	return ret
}

// GetPassword returns the decrypted password.
func (a *Auth) GetPassword() string {
	ret, err := crypto.Decrypt(crypto.MachineKey(), a.Pass)
	if err != nil {
		msg.Die("❌ Failed to decrypt pass: %s", err)
		return ""
	}

	return ret
}

// GetPassPhrase returns the decrypted passphrase.
func (a *Auth) GetPassPhrase() string {
	ret, err := crypto.Decrypt(crypto.MachineKey(), a.PassPhrase)
	if err != nil {
		msg.Die("❌ Failed to decrypt PassPhrase: %s", err)
		return ""
	}
	return ret
}

// SetUser encrypts and sets the username.
func (a *Auth) SetUser(user string) {
	if encryptedUser, err := crypto.Encrypt(crypto.MachineKey(), user); err != nil {
		msg.Die("❌ Failed to crypt user: %s", err)
	} else {
		a.User = encryptedUser
	}
}

// SetPass encrypts and sets the password.
func (a *Auth) SetPass(pass string) {
	if cPass, err := crypto.Encrypt(crypto.MachineKey(), pass); err != nil {
		msg.Die("❌ Failed to crypt pass: %s", err)
	} else {
		a.Pass = cPass
	}
}

// SetPassPhrase encrypts and sets the passphrase.
func (a *Auth) SetPassPhrase(passphrase string) {
	if cPassPhrase, err := crypto.Encrypt(crypto.MachineKey(), passphrase); err != nil {
		msg.Die("❌ Failed to crypt PassPhrase: %s", err)
	} else {
		a.PassPhrase = cPassPhrase
	}
}

// GetAuth returns the authentication method for a repository.
func (c *Configuration) GetAuth(repo string) transport.AuthMethod {
	auth := c.Auth[repo]

	switch {
	case auth == nil:
		return nil
	case auth.UseSSH:
		pem, err := os.ReadFile(auth.Path)
		if err != nil {
			msg.Die("❌ Failed to open ssh key %s", err)
		}
		var signer ssh.Signer

		if auth.GetPassPhrase() != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(pem, []byte(auth.GetPassPhrase()))
		} else {
			signer, err = ssh.ParsePrivateKey(pem)
		}

		if err != nil {
			msg.Die("❌ Failed to parse SSH private key: %v", err)
		}
		return &sshGit.PublicKeys{User: "git", Signer: signer}

	default:
		return &http.BasicAuth{Username: auth.GetUser(), Password: auth.GetPassword()}
	}
}

// SaveConfiguration saves the configuration to disk.
func (c *Configuration) SaveConfiguration() {
	jsonString, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		msg.Die("❌ Failed to parse config file", err.Error())
	}

	//nolint:gosec,nolintlint // Standard permissions for Boss cache directory
	err = os.MkdirAll(c.path, 0755)
	if err != nil {
		msg.Die("❌ Failed to create path", c.path, err.Error())
	}

	configPath := filepath.Join(c.path, consts.BossConfigFile)
	//nolint:gosec,nolintlint // Creating Boss configuration file in known location
	f, err := os.Create(configPath)
	if err != nil {
		msg.Die("❌ Failed to create file ", configPath, err.Error())
		return
	}

	defer f.Close()

	_, err = f.Write(jsonString)
	if err != nil {
		msg.Die("❌ Failed to write cache file", err.Error())
	}
}

// makeDefault creates a default configuration.
func makeDefault(configPath string) *Configuration {
	return &Configuration{
		path:                configPath,
		PurgeTime:           3,
		InternalRefreshRate: 5,
		LastInternalUpdate:  time.Now(),
		Auth:                make(map[string]*Auth),
		Key:                 crypto.Md5MachineID(),
		GitEmbedded:         true,
		GitShallow:          false, // Default to full clone for compatibility
	}
}

// LoadConfiguration loads the configuration from disk.
func LoadConfiguration(cachePath string) (*Configuration, error) {
	configuration := &Configuration{
		PurgeTime: 3,
	}

	configFileName := filepath.Join(cachePath, consts.BossConfigFile)
	//nolint:gosec,nolintlint // Reading Boss configuration file from cache directory
	buffer, err := os.ReadFile(configFileName)
	if err != nil {
		return makeDefault(cachePath), err
	}
	err = json.Unmarshal(buffer, configuration)
	if err != nil {
		msg.Err("❌ Failed to load cfg %s", err)
		return makeDefault(cachePath), err
	}
	if configuration.Key != crypto.Md5MachineID() {
		msg.Err("❌ Failed to load auth... recreate login accounts")
		configuration.Key = crypto.Md5MachineID()
		configuration.Auth = make(map[string]*Auth)
	}

	configuration.path = cachePath

	return configuration, nil
}
