package env

import (
	"encoding/json"
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils/crypto"
	"golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	sshGit "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var machineID = []byte(crypto.GetMachineID()[:16])

type Configuration struct {
	path                string
	Key                 string           `json:"id"`
	Auth                map[string]*Auth `json:"auth"`
	PurgeTime           int              `json:"purge_after"`
	InternalRefreshRate int              `json:"internal_refresh_rate"`
	LastPurge           time.Time        `json:"last_purge_cache"`
	LastInternalUpdate  time.Time        `json:"last_internal_update"`
	DelphiPath          string           `json:"delphi_path,omitempty"`
	ConfigVersion       int64            `json:"config_version"`
	GitEmbedded         bool             `json:"git_embedded"`
}

type Auth struct {
	UseSsh bool   `json:"use,omitempty"`
	Path   string `json:"path,omitempty"`
	User   string `json:"x,omitempty"`
	Pass   string `json:"y,omitempty"`
}

func (a *Auth) GetUser() string {
	if ret, err := crypto.Decrypt(machineID, a.User); err != nil {
		msg.Err("Fail to decrypt user.")
		return ""
	} else {
		return ret
	}
}

func (a *Auth) GetPassword() string {
	if ret, err := crypto.Decrypt(machineID, a.Pass); err != nil {
		msg.Err("Fail to decrypt pass.", err)
		return ""
	} else {
		return ret
	}
}

func (a *Auth) SetUser(user string) {
	if cUSer, err := crypto.Encrypt(machineID, user); err != nil {
		msg.Err("Fail to crypt user.", err)
	} else {
		a.User = cUSer
	}
}

func (a *Auth) SetPass(pass string) {
	if cPass, err := crypto.Encrypt(machineID, pass); err != nil {
		msg.Err("Fail to crypt pass.")
	} else {
		a.Pass = cPass
	}
}

func (c *Configuration) addAuth(repo string, auth *Auth) {
	if c.Auth == nil {
		c.Auth = make(map[string]*Auth)
	}
	c.Auth[repo] = auth
}

func (c *Configuration) removeAuth(repo string) {
	if c.Auth == nil {
		return
	}
	delete(c.Auth, repo)
}

func (c *Configuration) GetAuth(repo string) transport.AuthMethod {
	auth := c.Auth[repo]
	if auth == nil {
		return nil
	} else if auth.UseSsh {
		pem, e := ioutil.ReadFile(auth.Path)
		if e != nil {
			msg.Die("Fail to open ssh key %s", e)
		}
		var signer ssh.Signer

		signer, e = ssh.ParsePrivateKey(pem)

		if e != nil {
			panic(e)
		}
		return &sshGit.PublicKeys{User: "git", Signer: signer}

	} else {
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
	buffer, err := ioutil.ReadFile(configFileName)
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
