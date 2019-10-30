package builders

import (
	"os"

	"github.com/docker/docker/api/types/mount"
	"github.com/shibbybird/eazy-ci/lib/config"
	"github.com/shibbybird/eazy-ci/lib/utils"
)

type sbtEnvironmentBuilder struct{}

func (s sbtEnvironmentBuilder) GetBuildContainerOptions() (config.RuntimeConfig, error) {
	cacheMounts, err := s.GetLocalCacheMounts()

	if err != nil {
		return config.RuntimeConfig{}, err
	}

	pwd, err := os.Getwd()

	if err != nil {
		return config.RuntimeConfig{}, err
	}

	mounts := []mount.Mount{
		mount.Mount{
			Source:      pwd,
			Target:      "/root/build",
			Type:        mount.TypeBind,
			ReadOnly:    false,
			Consistency: mount.ConsistencyDelegated,
		},
	}

	mounts = append(mounts, cacheMounts...)

	return config.RuntimeConfig{
		User:        "root",
		Wait:        true,
		ExposePorts: false,
		Attach:      true,
		WorkingDir:  "/root/build",
		Mounts:      mounts,
	}, nil
}

func (s sbtEnvironmentBuilder) GetLocalCacheMounts() ([]mount.Mount, error) {
	homeDir, err := utils.GetEazyHomeDir()

	if err != nil {
		return nil, err
	}

	// Create a .sbt
	sbtDir := homeDir + "/.sbt"
	if _, err := os.Stat(sbtDir); os.IsNotExist(err) {
		os.Mkdir(sbtDir, 0775)
	}

	if err != nil {
		return nil, err
	}

	// Create ivy cache
	ivyCacheDir := homeDir + "/.ivy2"
	if _, err := os.Stat(ivyCacheDir); os.IsNotExist(err) {
		os.Mkdir(ivyCacheDir, 0775)
	}

	if err != nil {
		return nil, err
	}

	// Added /home/sbtuser/.ivy2 because of popular image: hseeberger/scala-sbt
	return []mount.Mount{
		mount.Mount{
			Source:      sbtDir,
			Target:      "/root/.sbt",
			Type:        mount.TypeBind,
			ReadOnly:    false,
			Consistency: mount.ConsistencyDelegated,
		},
		mount.Mount{
			Source:      sbtDir,
			Target:      "/home/sbtuser/.sbt",
			Type:        mount.TypeBind,
			ReadOnly:    false,
			Consistency: mount.ConsistencyDelegated,
		},
		mount.Mount{
			Source:      ivyCacheDir,
			Target:      "/root/.ivy2",
			Type:        mount.TypeBind,
			ReadOnly:    false,
			Consistency: mount.ConsistencyDelegated,
		},
		mount.Mount{
			Source:      ivyCacheDir,
			Target:      "/home/sbtuser/.ivy2",
			Type:        mount.TypeBind,
			ReadOnly:    false,
			Consistency: mount.ConsistencyDelegated,
		},
	}, nil
}
