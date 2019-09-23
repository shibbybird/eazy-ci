package utils

import (
	"errors"
	"strings"
	"testing"
)

func TestGetSSHAuth(t *testing.T) {
	_, err := GetSSHAuth("./testdata/does_not_exist")

	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Error(errors.New("Did not fail when reading file"))
	}

	key, err := GetSSHAuth("./testdata/id_rsa")

	if err != nil {
		t.Error(err)
	}

	if key.Name() != "ssh-public-keys" {
		t.Fail()
	}

	_, err = GetSSHAuth("./testdata/id_rsa_bad")

	if !strings.Contains(err.Error(), "ssh: no key found") {
		t.Error(errors.New("Should fail to parse file"))
	}
}
