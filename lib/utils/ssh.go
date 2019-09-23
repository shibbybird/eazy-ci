package utils

import (
	"io/ioutil"
	"log"
	"os/exec"

	"gopkg.in/src-d/go-git.v4/plumbing/transport"

	"golang.org/x/crypto/ssh"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

// SetUpSSHKeys to automatically add ssh keys by using ssh agent
func SetUpSSHKeys() error {
	cmd := exec.Command("ssh-add")
	err := cmd.Run()

	if err != nil {
		log.Println("Warn: ssh agent is not installed can not add ssh key")
		return err
	}

	return nil
}

// GetSSHAuth get ssh auth for git handshake
func GetSSHAuth(keyPath string) (transport.AuthMethod, error) {
	privateKeyFile, err := ioutil.ReadFile(keyPath)

	if err != nil {
		return nil, err
	}

	privateKey, err := ssh.ParsePrivateKey(privateKeyFile)

	if err != nil {
		return nil, err
	}

	auth := &ssh2.PublicKeys{
		User: "git", Signer: privateKey,
	}

	return auth, nil
}
