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

var supportedBuilders = map[string]buildEnvironment{
	"gradle": gradleEnvironmentBuilder{},
	"sbt":    sbtEnvironmentBuilder{},
}

func GetBuildEnvironment(envBuilder string) buildEnvironment {
	var builder buildEnvironment
	if b, ok := supportedBuilders[envBuilder]; ok {
		log.Println("Using " + envBuilder + " Build Environment")
		builder = b
	} else {
		log.Println("Using Default Build Environment")
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
