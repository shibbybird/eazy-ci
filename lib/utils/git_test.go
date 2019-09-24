package utils

import (
	"errors"
	"testing"
)

func TestParseGitURI(t *testing.T) {
	host, path, err := parseGitURI("github.com/shibbybird/eazy-ci")

	if err != nil {
		t.Error(err)
	}

	if host != "github.com" ||
		path != "shibbybird/eazy-ci" {
		t.Error(errors.New("Failed incorrect parse host: '" + host + "' path: '" + path + "'"))
	}
}

func TestFailParseGitURI(t *testing.T) {
	_, _, err := parseGitURI("github.comeazy-ci")

	if err == nil {
		t.Error(errors.New("Should have parsed URI"))
	}
}

func TestGenerateGitURL(t *testing.T) {
	host, path, err := parseGitURI("github.com/shibbybird/eazy-ci")

	if err != nil {
		t.Error(err)
	}

	url := generateGitURL(host, path)

	if url != "git@github.com:shibbybird/eazy-ci.git" {
		t.Error(errors.New("Incorrect url: '" + url + "'"))
	}
}
