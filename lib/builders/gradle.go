package builders

import (
	"os"

	"github.com/docker/docker/api/types/mount"
	"github.com/shibbybird/eazy-ci/lib/models"
	"github.com/shibbybird/eazy-ci/lib/utils"
)

type gradleEnvironmentBuilder struct{}

func (g gradleEnvironmentBuilder) GetBuildContainerOptions() (models.DockerConfig, error) {
	cacheMounts, err := g.GetLocalCacheMounts()

	if err != nil {
		return models.DockerConfig{}, err
	}

	pwd, err := os.Getwd()

	if err != nil {
		return models.DockerConfig{}, err
	}

	mounts := []mount.Mount{
		mount.Mount{
			Source:      pwd,
			Target:      "/root/build",
			Type:        mount.TypeBind,
			ReadOnly:    false,
			Consistency: mount.ConsistencyFull,
		},
	}

	mounts = append(mounts, cacheMounts...)

	return models.DockerConfig{
		User:        "root",
		Wait:        true,
		ExposePorts: false,
		Attach:      true,
		WorkingDir:  "/root/build",
		Mounts:      mounts,
	}, nil
}

func (g gradleEnvironmentBuilder) GetLocalCacheMounts() ([]mount.Mount, error) {
	homeDir, err := utils.GetEazyHomeDir()

	if err != nil {
		return nil, err
	}

	// Create a .gradleDir
	gradleDir := homeDir + "/.gradle"
	if _, err := os.Stat(gradleDir); os.IsNotExist(err) {
		os.Mkdir(gradleDir, 0775)
	}

	if err != nil {
		return nil, err
	}

	return []mount.Mount{
		mount.Mount{
			Source:      gradleDir,
			Target:      "/root/.gradle",
			Type:        mount.TypeBind,
			ReadOnly:    false,
			Consistency: mount.ConsistencyFull,
		},
		mount.Mount{
			Source:      gradleDir,
			Target:      "/home/gradle/.gradle",
			Type:        mount.TypeBind,
			ReadOnly:    false,
			Consistency: mount.ConsistencyFull,
		},
	}, nil
}
