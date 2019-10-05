package models

import (
	"gopkg.in/yaml.v2"
)

// EazyYml stuct for eazy.yml file in the repo's
type EazyYml struct {
	Name        string
	EazyVersion string `yaml:"eazyVersion"`
	Releases    []string
	Image       string
	Build       struct {
		Image   string
		Command []string
	}
	Deployment struct {
		Env    []string
		Ports  []string
		Health []string
	}
	Integration struct {
		Bootstrap        []string
		RunTest          []string `yaml:"runTest"`
		Dependencies     []string
		PeerDependencies []string `yaml:"peerDependencies"`
	}
}

func GetLatestImageName(eazy EazyYml) string {
	return eazy.Image + ":" + eazy.Releases[0]
}

func GetLatestIntegrationImageName(eazy EazyYml) string {
	return eazy.Image + "-integration:" + eazy.Releases[0]
}

// EazyzYmlUnmarshal EaztYml
func EazyYmlUnmarshal(in []byte) (EazyYml, error) {
	yml := EazyYml{}
	err := yaml.Unmarshal(in, &yml)
	return yml, err
}
