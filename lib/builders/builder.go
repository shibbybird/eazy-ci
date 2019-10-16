package builders

import (
	"log"
	"os"

	"github.com/docker/docker/api/types/mount"
	"github.com/shibbybird/eazy-ci/lib/config"
)

type buildEnvironment interface {
	GetBuildContainerOptions() (config.DockerConfig, error)
	GetLocalCacheMounts() ([]mount.Mount, error)
}

func GetBuildEnvironment(envBuilder string) buildEnvironment {
	var builder buildEnvironment

	switch envBuilder {
	case "gradle":
		log.Println("Using gradle build environment")
		builder = gradleEnvironmentBuilder{}
	case "sbt":
		log.Println("Using sbt build environment")
		builder = sbtEnvironmentBuilder{}
	default:
		log.Println("Your build environment does not have caching. Please contribute your build environment to Eazy!")
		builder = defaultEnvironmentBuilder{}
	}

	return builder
}

type defaultEnvironmentBuilder struct{}

func (g defaultEnvironmentBuilder) GetBuildContainerOptions() (config.DockerConfig, error) {
	pwd, err := os.Getwd()

	if err != nil {
		return config.DockerConfig{}, err
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

	return config.DockerConfig{
		User:        "root",
		Wait:        true,
		ExposePorts: false,
		Attach:      true,
		WorkingDir:  "/root/build",
		Mounts:      mounts,
	}, nil
}

func (g defaultEnvironmentBuilder) GetLocalCacheMounts() ([]mount.Mount, error) {
	return []mount.Mount{}, nil
}
