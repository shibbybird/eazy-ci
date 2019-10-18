package runtimes

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/go-connections/nat"
	"github.com/shibbybird/eazy-ci/lib/config"
)

// NewDockerRuntime returns a runtime for the docker runtime
func NewDockerRuntime() (ContainerRuntime, error) {
	dockerClient, err := client.NewClientWithOpts(client.WithVersion("1.40"))
	if err != nil {
		return DockerRuntime{}, err
	}

	return DockerRuntime{client: dockerClient}, nil
}

//StartContainerByEazyYml bootstrap runtime container using provided yaml config
func (runtime DockerRuntime) StartContainerByEazyYml(ctx context.Context, eazy config.EazyYml, imageOverride string, cfg config.RuntimeConfig, routableLinks *[]string, liveContainers *[]string) (string, error) {
	var image string

	if len(imageOverride) > 0 {
		image = imageOverride
	} else {
		image = config.GetLatestImageName(eazy)
	}

	if !cfg.SkipImagePull {
		reader, err := runtime.client.ImagePull(ctx, image, types.ImagePullOptions{})
		if err != nil {
			return "", err
		}

		aux := func(msg jsonmessage.JSONMessage) {
			var result types.Container
			if err := json.Unmarshal(*msg.Aux, &result); err != nil {
				log.Fatal(err)
			}
		}

		jsonmessage.DisplayJSONMessagesStream(reader, os.Stdout, os.Stdout.Fd(), true, aux)
		io.Copy(os.Stdout, reader)
	}

	containerID, err := createContainer(ctx, eazy, runtime, imageOverride, cfg, *routableLinks)
	if err != nil {
		return containerID, err
	}
	*liveContainers = append(*liveContainers, containerID)

	var oldStateIn *term.State
	if cfg.Attach {
		oldStateIn, _ = term.SetRawTerminal(os.Stdin.Fd())
	}

	err = startContainer(ctx, containerID, runtime, cfg)
	if err == nil {
		if cfg.IsRootImage {
			*routableLinks = append(*routableLinks, (containerID + ":" + eazy.Name))
		}
	}

	if cfg.Attach {
		term.RestoreTerminal(os.Stdin.Fd(), oldStateIn)
	}

	return containerID, err

}

func createContainer(ctx context.Context, eazy config.EazyYml, runtime DockerRuntime, imageOverride string, cfg config.RuntimeConfig, routableLinks []string) (string, error) {
	imageName := config.GetLatestImageName(eazy)

	if len(imageOverride) > 0 {
		imageName = imageOverride
	}

	pMap := nat.PortMap{}
	pSet := nat.PortSet{}
	if cfg.ExposePorts {
		for _, port := range eazy.Deployment.Ports {
			p := nat.Port(port + "/tcp")
			pMap[p] = []nat.PortBinding{{
				HostPort: port,
			}}
			pSet[p] = struct{}{}
		}
	}

	shouldOpenStdin := false
	if cfg.Attach {
		shouldOpenStdin = true
	}

	attach := cfg.Attach
	if cfg.Wait {
		attach = true
	}

	hostName := ""

	if cfg.IsRootImage {
		hostName = eazy.Name
	}

	response, err := runtime.client.ContainerCreate(ctx, &container.Config{
		User:         cfg.User,
		Hostname:     hostName,
		Domainname:   hostName,
		Image:        imageName,
		ExposedPorts: pSet,
		Env:          cfg.Env,
		Cmd:          cfg.Command,
		Tty:          attach,
		AttachStdin:  attach,
		AttachStdout: attach,
		AttachStderr: attach,
		OpenStdin:    shouldOpenStdin,
		WorkingDir:   cfg.WorkingDir,
	}, &container.HostConfig{
		Mounts:       cfg.Mounts,
		PortBindings: pMap,
		Links:        routableLinks,
	}, &network.NetworkingConfig{}, "")

	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func hijackConnection(ctx context.Context, resp types.HijackedResponse, attached bool) error {
	output := make(chan error)
	input := make(chan struct{})
	inErrCh := make(chan error)

	if attached {
		go func() {
			_, err := io.Copy(resp.Conn, os.Stdin)
			if _, ok := err.(term.EscapeError); ok {
				inErrCh <- err
			}
			resp.CloseWrite()
			close(input)
		}()
	}
	go func() {
		_, err := io.Copy(os.Stdout, resp.Reader)
		resp.Close()
		output <- err
	}()

	select {
	case err := <-output:
		return err
	case <-input:
		select {
		case err := <-output:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	case err := <-inErrCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func startContainer(ctx context.Context, containerID string, runtime DockerRuntime, cfg config.RuntimeConfig) error {
	var errResult error

	if cfg.Attach || cfg.Wait {
		resp, err := runtime.client.ContainerAttach(ctx, containerID, types.ContainerAttachOptions{
			Stream: true,
			Stdin:  cfg.Attach,
			Stdout: true,
			Stderr: true,
		})

		if err != nil {
			log.Println(err)
			return err
		}

		errCh := make(chan error, 1)

		go func() {
			errCh <- func() error {
				return hijackConnection(ctx, resp, cfg.Attach)
			}()
		}()

		defer resp.Close()
		defer resp.CloseWrite()
	}

	err := runtime.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	if cfg.Wait {

		statusResult := make(chan int)
		chn, errCh := runtime.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
		go func() {
			select {
			case <-errCh:
				statusResult <- 7
			case out := <-chn:
				statusResult <- int(out.StatusCode)
			}
		}()

		statusCode := <-statusResult

		if statusCode > 0 {
			errResult = errors.New("Error Starting Container - Status Code: " + string(statusCode))
		}

	}

	return errResult
}

// BuildAndRunContainer builds given container based on config and starts it
func (runtime DockerRuntime) BuildAndRunContainer(ctx context.Context, eazy config.EazyYml, cfg config.RuntimeConfig, routableLinks *[]string, liveContainers *[]string) (string, error) {

	tar, err := archive.TarWithOptions("./", &archive.TarOptions{})
	if err != nil {
		return "", err
	}

	defer tar.Close()

	opt := types.ImageBuildOptions{
		Dockerfile: cfg.Dockerfile,
	}

	resp, err := runtime.client.ImageBuild(ctx, tar, opt)
	if err != nil {
		return "", err
	}

	imageID := ""
	aux := func(msg jsonmessage.JSONMessage) {
		var result types.BuildResult
		if err := json.Unmarshal(*msg.Aux, &result); err != nil {
			log.Fatal(err)
		} else {
			imageID = result.ID
		}
	}

	jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stdout, os.Stdout.Fd(), true, aux)

	if err == nil {
		resp.Body.Close()
	} else {
		return "", err
	}

	containerID, err := createContainer(ctx, eazy, runtime, imageID, cfg, *routableLinks)
	if err != nil {
		return containerID, err
	}

	*liveContainers = append(*liveContainers, containerID)

	var oldStateIn *term.State
	if cfg.Attach {
		oldStateIn, _ = term.SetRawTerminal(os.Stdin.Fd())
	}

	err = startContainer(ctx, containerID, runtime, cfg)
	if err == nil {
		if cfg.IsRootImage {
			*routableLinks = append(*routableLinks, (containerID + ":" + eazy.Name))
		}
	}

	if cfg.Attach {
		term.RestoreTerminal(os.Stdin.Fd(), oldStateIn)
	}

	return imageID, err

}

// KillContainer stops a runtime container
func (runtime DockerRuntime) KillContainer(ctx context.Context, id string) error {
	return runtime.client.ContainerKill(ctx, id, "KILL")
}
