package env

import (
	"errors"
	"fmt"
	"os"

	"github.com/hashload/boss/pkg/msg"
	"github.com/pterm/pterm"
	"golang.org/x/crypto/ssh"
)

func getSSHPass(auth *Auth) (string, error) {
	msg.Warn("Passphrase is required for the private key %s", auth.Path)
	description := "Enter passphrase for key " + auth.Path
	pass, err := pterm.DefaultInteractiveTextInput.WithMask("•").Show(description)
	if err != nil {
		return "", fmt.Errorf("error on get pass: %w", err)
	}

	res, err := pterm.DefaultInteractiveConfirm.Show(
		"Do you want to save passphrase in configuration?",
	)
	if err != nil {
		return "", fmt.Errorf("error on passphrase storage: %w", err)
	}

	if res {
		auth.SetPassPhrase(pass)
	}

	return pass, nil
}

func getSigner(auth *Auth) (ssh.Signer, error) {
	pem, err := os.ReadFile(auth.Path)
	if err != nil {
		return nil, fmt.Errorf("fail to open ssh key %w", err)
	}

	var signer ssh.Signer
	if auth.GetPassPhrase() != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(pem, []byte(auth.GetPassPhrase()))
	} else {
		signer, err = ssh.ParsePrivateKey(pem)
	}

	if err != nil && errors.Is(err, &ssh.PassphraseMissingError{}) {
		pass, err := getSSHPass(auth)
		if err != nil {
			return nil, err
		}

		signer, err = ssh.ParsePrivateKeyWithPassphrase(pem, []byte(pass))
	}

	if err != nil {
		return nil, fmt.Errorf("fail to parse ssh key %w", err)
	}

	return signer, nil
}
