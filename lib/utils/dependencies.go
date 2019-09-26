package utils

import (
	"github.com/shibbybird/eazy-ci/lib/models"
)

func GetDependencies(in models.EazyYml, out *[]models.EazyYml, sshKeyPath string) error {
	for _, d := range in.Integration.Dependencies {
		data, err := GetEazyYmlFromRepository(d, sshKeyPath)
		if err != nil {
			return err
		}
		eazy, err := models.EazyYmlUnmarshal(data)
		if err != nil {
			return err
		}
		if len(eazy.Integration.Dependencies) > 0 {
			err = GetDependencies(eazy, out, sshKeyPath)
			if err != nil {
				return err
			}
		}
		*out = append(*out, eazy)
	}
	return nil
}

func GetPeerDependencies(in models.EazyYml, out *[]models.EazyYml, peers map[string]bool, sshKeyPath string) error {
	for _, d := range in.Integration.PeerDependencies {
		data, err := GetEazyYmlFromRepository(d, sshKeyPath)
		if err != nil {
			return err
		}
		eazy, err := models.EazyYmlUnmarshal(data)
		if err != nil {
			return err
		}
		if len(eazy.Integration.PeerDependencies) > 0 {
			err = GetPeerDependencies(eazy, out, peers, sshKeyPath)
			if err != nil {
				return err
			}
		}
		if _, ok := peers[d]; !ok {
			*out = append(*out, eazy)
			peers[d] = true
		}
	}
	return nil
}
