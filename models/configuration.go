package models

import (
	"encoding/json"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils/crypto"
	"golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io/ioutil"
	"os"
	"path/filepath"

	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var machineID = []byte(crypto.GetMachineID()[:16])

type Configuration struct {
	Key        string           `json:"id"`
	Auth       map[string]*Auth `json:"auth"`
	PurgeTime  int              `json:"purgeAfter"`
	LastPurge  string           `json:"last_purge"`
	DelphiPath string           `json:"delphi_path,omitempty"`
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
		signer, e := ssh.ParsePrivateKey(pem)
		if e != nil {
			panic(e)
		}
		return &ssh2.PublicKeys{User: "git", Signer: signer}

	} else {
		return &http.BasicAuth{Username: auth.GetUser(), Password: auth.GetPassword()}
	}
}

func (c *Configuration) SaveConfiguration() error {
	d, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = os.MkdirAll(env.GetBossHome(), 0755)
	if err != nil {
		return err
	}

	p := filepath.Join(env.GetBossHome(), "boss.cfg.json")
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(d)
	return err
}

func defaultCreate() *Configuration {
	return &Configuration{
		PurgeTime: 3,
		Auth:      make(map[string]*Auth),
		Key:       crypto.Md5MachineID(),
	}
}

var GlobalConfiguration, _ = loadConfiguration()

func loadConfiguration() (*Configuration, error) {

	c := &Configuration{
		PurgeTime: 3,
	}
	p := filepath.Join(env.GetBossHome(), "boss.cfg.json")
	f, err := ioutil.ReadFile(p)
	if err != nil {
		return defaultCreate(), err
	}
	err = json.Unmarshal(f, c)
	if err != nil {
		msg.Err("Fail to load cfg %s", err)
		return defaultCreate(), err
	}
	if c.Key != crypto.Md5MachineID() {
		msg.Err("Falied to load auths... recreate logins")
		c.Key = crypto.Md5MachineID()
		c.Auth = make(map[string]*Auth)
	}

	return c, nil
}
