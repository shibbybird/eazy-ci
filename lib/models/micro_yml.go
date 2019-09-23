package models

import (
	"gopkg.in/yaml.v2"
)

// MicroYml stuct for micro.yml file in the repo's
type MicroYml struct {
	MicroVersion string `yaml:"microVersion"`
	Releases     []string
	Image        string
	Deployment   struct {
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

// MicroYmlUnmarshal MicroYml
func MicroYmlUnmarshal(in []byte) (MicroYml, error) {
	yml := MicroYml{}
	err := yaml.Unmarshal(in, &yml)
	return yml, err
}
