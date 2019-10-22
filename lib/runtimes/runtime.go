package runtimes

import (
	"github.com/shibbybird/eazy-ci/lib/config"
)

const (
	dockerRuntime = "docker"
)

// NewRuntime returns a container runtime
func NewRuntime(cfg config.EazyYml) (ContainerRuntime, error) {
	switch cfg.Runtime {
	case dockerRuntime:
		return NewDockerRuntime()
	default:
		// return nil, errors.New("No Runtime Defined")
		return NewDockerRuntime()
	}
}
