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

	Advices struct {
		SetupPath bool `json:"setup_path,omitempty"`
	} `json:"advices"`
}

type Auth struct {
	UseSSH     bool   `json:"use,omitempty"`
	Path       string `json:"path,omitempty"`
	User       string `json:"user,omitempty"`
	Pass       string `json:"pass,omitempty"`
	PassPhrase string `json:"keypass,omitempty"`
}

func (a *Auth) GetUser() string {
	ret, err := crypto.Decrypt(crypto.MachineKey(), a.User)
	if err != nil {
		msg.Err("Fail to decrypt user.")
		return ""
	}
	return ret
}

func (a *Auth) GetPassword() string {
	ret, err := crypto.Decrypt(crypto.MachineKey(), a.Pass)
	if err != nil {
		msg.Err("Fail to decrypt pass.", err)
		return ""
	}

	return ret
}

func (a *Auth) GetPassPhrase() string {
	ret, err := crypto.Decrypt(crypto.MachineKey(), a.PassPhrase)
	if err != nil {
		msg.Err("Fail to decrypt PassPhrase.", err)
		return ""
	}
	return ret
}

func (a *Auth) SetUser(user string) {
	if encryptedUser, err := crypto.Encrypt(crypto.MachineKey(), user); err != nil {
		msg.Err("Fail to crypt user.", err)
	} else {
		a.User = encryptedUser
	}
}

func (a *Auth) SetPass(pass string) {
	if cPass, err := crypto.Encrypt(crypto.MachineKey(), pass); err != nil {
		msg.Err("Fail to crypt pass.")
	} else {
		a.Pass = cPass
	}
}

func (a *Auth) SetPassPhrase(passphrase string) {
	if cPassPhrase, err := crypto.Encrypt(crypto.MachineKey(), passphrase); err != nil {
		msg.Err("Fail to crypt PassPhrase.")
	} else {
		a.PassPhrase = cPassPhrase
	}
}

func (c *Configuration) GetAuth(repo string) transport.AuthMethod {
	auth := c.Auth[repo]

	switch {
	case auth == nil:
		return nil
	case auth.UseSSH:
		pem, err := os.ReadFile(auth.Path)
		if err != nil {
			msg.Die("Fail to open ssh key %s", err)
		}
		var signer ssh.Signer

		if auth.GetPassPhrase() != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(pem, []byte(auth.GetPassPhrase()))
		} else {
			signer, err = ssh.ParsePrivateKey(pem)
		}

		if err != nil {
			panic(err)
		}
		return &sshGit.PublicKeys{User: "git", Signer: signer}

	default:
		return &http.BasicAuth{Username: auth.GetUser(), Password: auth.GetPassword()}
	}
}

func (c *Configuration) SaveConfiguration() {
	jsonString, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		msg.Die("Error on parse config file", err.Error())
	}

	err = os.MkdirAll(c.path, 0755)
	if err != nil {
		msg.Die("Failed on create path", c.path, err.Error())
	}

	configPath := filepath.Join(c.path, consts.BossConfigFile)
	f, err := os.Create(configPath)
	if err != nil {
		msg.Die("Failed on create file ", configPath, err.Error())
		return
	}

	defer f.Close()

	_, err = f.Write(jsonString)
	if err != nil {
		msg.Die("Failed on write cache file", err.Error())
	}
}

func makeDefault(configPath string) *Configuration {
	return &Configuration{
		path:                configPath,
		PurgeTime:           3,
		InternalRefreshRate: 5,
		LastInternalUpdate:  time.Now(),
		Auth:                make(map[string]*Auth),
		Key:                 crypto.Md5MachineID(),
		GitEmbedded:         true,
	}
}

func LoadConfiguration(cachePath string) (*Configuration, error) {
	configuration := &Configuration{
		PurgeTime: 3,
	}

	configFileName := filepath.Join(cachePath, consts.BossConfigFile)
	buffer, err := os.ReadFile(configFileName)
	if err != nil {
		return makeDefault(cachePath), err
	}
	err = json.Unmarshal(buffer, configuration)
	if err != nil {
		msg.Err("Fail to load cfg %s", err)
		return makeDefault(cachePath), err
	}
	if configuration.Key != crypto.Md5MachineID() {
		msg.Err("Failed to load auth... recreate login accounts")
		configuration.Key = crypto.Md5MachineID()
		configuration.Auth = make(map[string]*Auth)
	}

	configuration.path = cachePath

	return configuration, nil
}
