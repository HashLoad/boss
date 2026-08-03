package env_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashload/boss/pkg/env"
	"golang.org/x/crypto/ssh"
)

// A cache cloned over HTTPS keeps its HTTPS remote even after `boss login -s`
// configures SSH for the host. Handing the SSH credential to the HTTP transport
// makes go-git reject the fetch with "invalid auth method", so every dependency
// of that host silently stops updating.
func TestGetAuthForURLDropsSSHCredentialOnHTTPRemote(t *testing.T) {
	config := newConfigWithAuth(t, &env.Auth{UseSSH: true, Path: "/home/user/.ssh/id_ed25519"})

	if got := config.GetAuthForURL("github.com", "https://github.com/hashload/horse"); got != nil {
		t.Errorf("GetAuthForURL() = %v, want nil for an SSH credential on an HTTPS remote", got)
	}
}

func TestGetAuthForURLDropsBasicCredentialOnSSHRemote(t *testing.T) {
	auth := &env.Auth{}
	auth.SetUser("octocat")
	auth.SetPass("s3cr3t")
	config := newConfigWithAuth(t, auth)

	if got := config.GetAuthForURL("github.com", "git@github.com:hashload/horse"); got != nil {
		t.Errorf("GetAuthForURL() = %v, want nil for a basic credential on an SSH remote", got)
	}
}

func TestGetAuthForURLKeepsMatchingCredential(t *testing.T) {
	auth := &env.Auth{}
	auth.SetUser("octocat")
	auth.SetPass("s3cr3t")
	config := newConfigWithAuth(t, auth)

	got := config.GetAuthForURL("github.com", "https://github.com/hashload/horse")
	basic, ok := got.(*http.BasicAuth)
	if !ok {
		t.Fatalf("GetAuthForURL() = %T, want *http.BasicAuth", got)
	}
	if basic.Username != "octocat" || basic.Password != "s3cr3t" {
		t.Errorf("GetAuthForURL() returned %q/%q, want octocat/s3cr3t", basic.Username, basic.Password)
	}
}

func TestGetAuthForURLWithoutStoredCredential(t *testing.T) {
	config := newConfigWithAuth(t, nil)

	if got := config.GetAuthForURL("gitlab.com", "https://gitlab.com/group/project"); got != nil {
		t.Errorf("GetAuthForURL() = %v, want nil when nothing is stored for the host", got)
	}
}

// An empty URL means the caller could not determine the transport. Falling back
// to the stored credential keeps the previous behaviour rather than silently
// dropping authentication.
func TestGetAuthForURLFallsBackWhenURLIsUnknown(t *testing.T) {
	auth := &env.Auth{}
	auth.SetUser("octocat")
	auth.SetPass("s3cr3t")
	config := newConfigWithAuth(t, auth)

	if got := config.GetAuthForURL("github.com", ""); got == nil {
		t.Error("GetAuthForURL() = nil, want the stored credential when the URL is unknown")
	}
}

func TestGetAuthForURLTransportDetection(t *testing.T) {
	sshAuth := &env.Auth{UseSSH: true, Path: writeTestSSHKey(t)}
	sshAuth.SetPassPhrase(testKeyPassphrase)
	config := newConfigWithAuth(t, sshAuth)

	sshRemotes := []string{
		"git@github.com:hashload/horse",
		"ssh://git@github.com/hashload/horse",
		"git@mygitlab.domain.de:delphi/libraries/mylib.git",
	}
	for _, remote := range sshRemotes {
		if got := config.GetAuthForURL("github.com", remote); got == nil {
			t.Errorf("GetAuthForURL(%q) = nil, want the SSH credential", remote)
		}
	}

	httpRemotes := []string{
		"https://github.com/hashload/horse",
		"http://github.com/hashload/horse",
		// A user in an HTTPS URL must not be mistaken for scp-like syntax.
		"https://octocat@github.com/hashload/horse",
	}
	for _, remote := range httpRemotes {
		if got := config.GetAuthForURL("github.com", remote); got != nil {
			t.Errorf("GetAuthForURL(%q) = %v, want nil", remote, got)
		}
	}
}

// testKeyPassphrase protects the generated key. Using an encrypted key keeps
// this test independent of how an empty passphrase is handled.
const testKeyPassphrase = "test-passphrase"

// writeTestSSHKey writes an encrypted ed25519 key and returns its path.
// The SSH branch of GetAuth parses the key file, so it needs a real one.
func writeTestSSHKey(t *testing.T) string {
	t.Helper()

	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate the test key: %v", err)
	}

	block, err := ssh.MarshalPrivateKeyWithPassphrase(private, "", []byte(testKeyPassphrase))
	if err != nil {
		t.Fatalf("Failed to marshal the test key: %v", err)
	}

	path := filepath.Join(t.TempDir(), "id_ed25519")
	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0600); err != nil {
		t.Fatalf("Failed to write the test key: %v", err)
	}
	return path
}

func newConfigWithAuth(t *testing.T, auth *env.Auth) *env.Configuration {
	t.Helper()

	config, err := env.LoadConfiguration(t.TempDir())
	if config == nil {
		t.Fatalf("LoadConfiguration() returned no configuration: %v", err)
	}
	if auth != nil {
		config.Auth["github.com"] = auth
	}
	return config
}
