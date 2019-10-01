package utils

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/shibbybird/eazy-ci/lib/models"
)

func StartContainerByEazyYml(ctx context.Context, eazy models.EazyYml, commands []string, shouldBlock bool, isHostMode bool, exposePorts bool, imageOverride string) (string, error) {

	dockerClient, err := client.NewClientWithOpts(client.WithVersion("1.40"))
	if err != nil {
		return "", err
	}

	reader, err := dockerClient.ImagePull(ctx, models.GetLatestImageName(eazy), types.ImagePullOptions{})
	if err != nil {
		return "", err
	}

	io.Copy(os.Stdout, reader)

	containerID, err := createContainer(ctx, eazy, dockerClient, commands, isHostMode, exposePorts, imageOverride, nil)
	if err != nil {
		return containerID, err
	}

	err = startContainer(ctx, containerID, dockerClient, shouldBlock)

	return containerID, err

}

func createContainer(ctx context.Context, eazy models.EazyYml, dockerClient *client.Client, commands []string, isHostMode bool, exposePorts bool, imageOverride string, environmentArr []string) (string, error) {
	imageName := models.GetLatestImageName(eazy)

	if len(imageOverride) > 0 {
		imageName = imageOverride
	}

	pMap := nat.PortMap{}
	pSet := nat.PortSet{}
	if exposePorts {
		for _, port := range eazy.Deployment.Ports {
			p := nat.Port(port + "/tcp")
			pMap[p] = []nat.PortBinding{{
				HostPort: port,
			}}
			pSet[p] = struct{}{}
		}
	}

	var networkMode container.NetworkMode

	if isHostMode {
		networkMode = "host"
	} else {
		networkMode = ""
	}

	response, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		ExposedPorts: pSet,
		Env:          environmentArr,
		Cmd:          commands,
	}, &container.HostConfig{
		NetworkMode:  networkMode,
		PortBindings: pMap,
	}, &network.NetworkingConfig{}, "")

	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func startContainer(ctx context.Context, containerID string, dockerClient *client.Client, shouldBlock bool) error {
	err := dockerClient.ContainerStart(ctx, containerID, types.ContainerStartOptions{})

	if err != nil {
		return err
	}

	if shouldBlock {
		chn, errCh := dockerClient.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
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

		out, _ := dockerClient.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
		})

		io.Copy(os.Stdout, out)
	}

	return err
}

func BuildAndRunContainer(ctx context.Context, environmentArr []string, eazy models.EazyYml, dockerfilePath string, commands []string, shouldBlock bool, isHostMode bool, exposePorts bool) (string, error) {

	dockerClient, err := client.NewClientWithOpts(client.WithVersion("1.40"))
	tar, err := archive.TarWithOptions("./", &archive.TarOptions{})
	if err != nil {
		return "", err
	}

	defer tar.Close()

	opt := types.ImageBuildOptions{
		Dockerfile: dockerfilePath,
	}

	resp, err := dockerClient.ImageBuild(ctx, tar, opt)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	tee := io.TeeReader(resp.Body, &buffer)

	io.Copy(os.Stdout, tee)

	respBytes, err := ioutil.ReadAll(&buffer)
	if err != nil {
		return "", err
	}

	responseStr := string(respBytes)
	idx := strings.Index(responseStr, "Successfully built")

	var imageID string
	if idx > 0 {
		imageID = responseStr[(idx + len("Successfully build") + 1):(idx + len("Successfully build") + 1 + 12)]
	}

	if err == nil {
		resp.Body.Close()
	} else {
		return "", err
	}

	containerID, err := createContainer(ctx, eazy, dockerClient, commands, isHostMode, exposePorts, imageID, environmentArr)
	if err != nil {
		return containerID, err
	}

	err = startContainer(ctx, containerID, dockerClient, shouldBlock)

	return containerID, err

}

func KillContainer(ctx context.Context, id string) error {
	dockerClient, err := client.NewClientWithOpts(client.WithVersion("1.40"))
	if err != nil {
		return err
	}
	err = dockerClient.ContainerKill(ctx, id, "KILL")
	return err
}
