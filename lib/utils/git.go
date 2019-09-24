package utils

import (
	"errors"
	"io/ioutil"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/client"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

func CreateGitClient() {
	httpClient := GetHttpClient()
	client.InstallProtocol("https", githttp.NewClient(httpClient))
}

// parseGitUri return host and path
func parseGitURI(uri string) (string, string, error) {
	i := strings.Index(uri, "/")
	if i > -1 {
		return uri[:i], uri[i+1:], nil
	} else {
		return "", "", errors.New("Failed to parse git uri e.g. 'github.com/shibbybird/eazy-ci'")
	}
}

func generateGitURL(host string, path string) string {
	return "git@" + host + ":" + path + ".git"
}

func GetRepository(uri string, sshKey string) (string, error) {
	dir, err := ioutil.TempDir("", "eazyci")
	if err != nil {
		return "", err
	}

	var auth transport.AuthMethod

	if len(sshKey) > 0 {
		auth, err = GetSSHAuth(sshKey)

		if err != nil {
			return "", err
		}
	} else {
		auth = nil
	}

	host, path, err := parseGitURI(uri)
	if err != nil {
		return "", err
	}

	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL:  generateGitURL(host, path),
		Auth: auth,
	})

	return dir, err
}

func GetEazyYmlFromRepository(uri string, ssKey string) ([]byte, error) {
	dir, err := GetRepository(uri, ssKey)

	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(dir + "/eazy.yml")

	return data, err
}
