package models

import (
	"gopkg.in/yaml.v2"
)

// EazyYml stuct for eazy.yml file in the repo's
type EazyYml struct {
	EazyVersion string `yaml:"eazyVersion"`
	Releases    []string
	Image       string
	Deployment  struct {
		Ports  []int
		Health []string
	}
	Integration struct {
		Bootstrap        []string
		RunTest          []string `yaml:"runTest"`
		Dependencies     []string
		PeerDependencies []string `yaml:"peerDependencies"`
	}
}

// EazyzYmlUnmarshal EaztYml
func EazyYmlUnmarshal(in []byte) (EazyYml, error) {
	yml := EazyYml{}
	err := yaml.Unmarshal(in, &yml)
	return yml, err
}

func GetEazyYmlDependencies(in EazyYml, out *[]EazyYml) error {

	/*
		for _, dep := range in.Integration.Dependencies {

		}
	*/

	return nil
}
