package runtimes

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/shibbybird/eazy-ci/lib/config"
)

// ContainerRuntime interface for contianer runtimes
type ContainerRuntime interface {
	StartContainerByEazyYml(ctx context.Context, eazy config.EazyYml, imageOverride string, cfg config.RuntimeConfig, routableLinks *[]string, liveContainers *[]string) (string, error)
	BuildAndRunContainer(ctx context.Context, eazy config.EazyYml, cfg config.RuntimeConfig, routableLinks *[]string, liveContainers *[]string) (string, error)
	KillContainer(ctx context.Context, id string) error
}

//DockerRuntime is an execution client for docker
type DockerRuntime struct {
	client *client.Client
}
