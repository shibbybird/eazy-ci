package utils

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/shibbybird/eazy-ci/lib/models"
)

func StartContainerByEazyYml(eazy models.EazyYml, commands []string, shouldBlock bool, isHostMode bool, ctx context.Context) (string, error) {

	cli, err := client.NewClientWithOpts(client.WithVersion("1.40"))
	if err != nil {
		return "", err
	}

	reader, err := cli.ImagePull(ctx, models.GetLatestImageName(eazy), types.ImagePullOptions{})
	if err != nil {
		return "", err
	}

	io.Copy(os.Stdout, reader)

	containerID, err := createContainer(eazy, cli, commands, isHostMode, ctx)
	if err != nil {
		return containerID, err
	}

	err = startContainer(containerID, cli, shouldBlock, ctx)

	return containerID, err

}

func createContainer(eazy models.EazyYml, dockerCli *client.Client, commands []string, isHostMode bool, ctx context.Context) (string, error) {
	imageName := models.GetLatestImageName(eazy)

	pMap := nat.PortMap{}
	pSet := nat.PortSet{}

	for _, port := range eazy.Deployment.Ports {
		p := nat.Port(string(port) + "/tcp")
		pMap[p] = []nat.PortBinding{{
			HostPort: string(port),
		}}
		pSet[p] = struct{}{}
	}

	var networkMode container.NetworkMode

	if isHostMode {
		networkMode = "host"
	} else {
		networkMode = ""
	}

	response, err := dockerCli.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		ExposedPorts: pSet,
	}, &container.HostConfig{
		NetworkMode:  networkMode,
		PortBindings: pMap,
	}, &network.NetworkingConfig{}, "")

	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func startContainer(containerID string, dockerCli *client.Client, shouldBlock bool, ctx context.Context) error {
	err := dockerCli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})

	if err != nil {
		return err
	}

	if shouldBlock {
		chn, errCh := dockerCli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
		case out := <-chn:
			statusCode := int(out.StatusCode)
			if statusCode > 0 {
				err = errors.New("Error Starting Container - Status Code: " + string(statusCode))
			}
		}

		out, _ := dockerCli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
		})

		io.Copy(os.Stdout, out)
	}

	return err
}
