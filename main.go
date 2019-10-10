package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/docker/docker/pkg/term"
	"github.com/shibbybird/eazy-ci/lib/builders"
	"github.com/shibbybird/eazy-ci/lib/utils"

	"github.com/shibbybird/eazy-ci/lib/models"
)

var VERSION = "v0.0.2"

var liveContainerIDs = []string{}
var routableLinks = []string{}

var oldStateOut *term.State = nil

// end of code for environment variables

func main() {
	ctx := context.Background()

	// Create a .eazy directory in user home
	homeDir, err := utils.GetEazyHomeDir()
	if err != nil {
		fail(ctx, err)
	}
	if _, err := os.Stat(homeDir); os.IsNotExist(err) {
		os.Mkdir(homeDir, 0775)
	}

	oldStateOut, _ = term.SetRawTerminalOutput(os.Stdout.Fd())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			cleanUp(ctx, 1, nil)
		}
	}()

	getVersion := flag.Bool("v", false, "Get version info")
	filePath := flag.String("f", "./eazy.yml", "The Eazy CI file")
	isDev := flag.Bool("d", false, "Run dependencies and peer depedencies")
	openPortsLocally := flag.Bool("p", false, "Open ports to depedencies and project containers locally. DISCLAIMER: If there are port conflicts starting eazy-ci will fail.")
	isIntegration := flag.Bool("i", false, "Run dependencies, peer dependencies, and build/start Dockerfile")
	pemKeyPath := flag.String("k", "", "File path for ssh private key for github access")

	flag.Parse()

	if *getVersion {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	fileData, err := ioutil.ReadFile(*filePath)
	if err != nil {
		fail(ctx, err)
	}

	yml, err := models.EazyYmlUnmarshal(fileData)
	if err != nil {
		fail(ctx, err)
	}

	dependencies := []models.EazyYml{}

	err = utils.GetDependencies(yml, &dependencies, *pemKeyPath)

	// try to set up ssh agent if ssh isn't working
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "ssh") {
			err = utils.SetUpSSHKeys()
			if err != nil {
				fail(ctx, err)
			}
			err = utils.GetDependencies(yml, &dependencies, *pemKeyPath)
			if err != nil {
				fail(ctx, err)
			}
		} else {
			fail(ctx, err)
		}
	}

	peerDependencies := []models.EazyYml{}
	peerDependenciesSet := map[string]bool{}

	// Collect Peer Dependencies
	for _, d := range dependencies {
		err = utils.GetPeerDependencies(d, &peerDependencies, peerDependenciesSet, *pemKeyPath)
		if err != nil {
			fail(ctx, errors.New("can not find all peer dependencies"))
		}
	}
	err = utils.GetPeerDependencies(yml, &peerDependencies, peerDependenciesSet, *pemKeyPath)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "ssh") {
			err = utils.SetUpSSHKeys()
			if err != nil {
				fail(ctx, err)
			}
			err = utils.GetPeerDependencies(yml, &peerDependencies, peerDependenciesSet, *pemKeyPath)
			if err != nil {
				fail(ctx, errors.New("can not find peer dependencies on eazy.yml"))
			}
		} else {
			fail(ctx, err)
		}
	}

	for _, d := range peerDependencies {
		startUnit(ctx, d, *openPortsLocally)
	}

	for _, d := range dependencies {
		startUnit(ctx, d, *openPortsLocally)
	}

	envBuilder := builders.GetBuildEnvironment(yml.Build.BuildEnvironment)
	localCacheMounts, err := envBuilder.GetLocalCacheMounts()

	if err != nil {
		fail(ctx, err)
	}

	buildImageDocker, err := envBuilder.GetBuildContainerOptions()

	if err != nil {
		fail(ctx, err)
	}

	if len(yml.Integration.Bootstrap) > 0 {
		_, err := utils.BuildAndRunContainer(ctx, yml, models.DockerConfig{
			Dockerfile:  "Integration.Dockerfile",
			Command:     yml.Integration.Bootstrap,
			Wait:        true,
			ExposePorts: false,
			Attach:      false,
			Mounts:      localCacheMounts,
		}, &routableLinks, &liveContainerIDs)

		if err != nil {
			fail(ctx, err)
		}
	}

	if !*isDev {

		if len(yml.Build.Image) > 0 {
			buildImageDocker.Command = yml.Build.Command
			_, err := utils.StartContainerByEazyYml(ctx, yml, yml.Build.Image,
				buildImageDocker, &routableLinks, &liveContainerIDs)

			if err != nil {
				fail(ctx, err)
			}
		}

		_, err = utils.BuildAndRunContainer(ctx, yml, models.DockerConfig{
			Env:         yml.Deployment.Env,
			Dockerfile:  "Dockerfile",
			Command:     []string{},
			Wait:        false,
			ExposePorts: *openPortsLocally,
			Attach:      false,
			IsRootImage: true,
		}, &routableLinks, &liveContainerIDs)

		if err != nil {
			fail(ctx, err)
		}

		if len(yml.Deployment.Health) > 0 {

			_, err := utils.BuildAndRunContainer(ctx, yml, models.DockerConfig{
				Dockerfile:  "Integration.Dockerfile",
				Command:     yml.Deployment.Health,
				Wait:        true,
				ExposePorts: false,
				Attach:      false,
				Mounts:      localCacheMounts,
			}, &routableLinks, &liveContainerIDs)

			if err != nil {
				fail(ctx, err)
			}
		}
	}

	if *isDev || *isIntegration {
		buildImageDocker.Command = []string{"/bin/bash"}

		// If you have a build image then use this for dev
		// if not then use the integration docker image
		// Why do you not need a build image?
		if len(yml.Build.Image) > 0 {
			_, err = utils.StartContainerByEazyYml(ctx, yml, yml.Build.Image, buildImageDocker, &routableLinks, &liveContainerIDs)
		} else {
			buildImageDocker.Dockerfile = "Integration.Dockerfile"
			_, err = utils.BuildAndRunContainer(ctx, yml, buildImageDocker, &routableLinks, &liveContainerIDs)
		}

		if err != nil {
			fail(ctx, err)
		}

	} else {
		_, err := utils.BuildAndRunContainer(ctx, yml, models.DockerConfig{
			Dockerfile:  "Integration.Dockerfile",
			Command:     yml.Integration.RunTest,
			Wait:        true,
			ExposePorts: false,
			Attach:      false,
			Mounts:      localCacheMounts,
		}, &routableLinks, &liveContainerIDs)

		if err != nil {
			fail(ctx, err)
		}
		success(ctx)
	}

	success(ctx)
}

func startUnit(ctx context.Context, yml models.EazyYml, openPortsLocally bool) {
	if len(yml.Integration.Bootstrap) > 0 {
		_, err := utils.StartContainerByEazyYml(ctx, yml, models.GetLatestIntegrationImageName(yml), models.DockerConfig{
			Command:     yml.Integration.Bootstrap,
			Wait:        true,
			ExposePorts: false,
		}, &routableLinks, &liveContainerIDs)

		if err != nil {
			fail(ctx, err)
		}
	}
	_, err := utils.StartContainerByEazyYml(ctx, yml, "", models.DockerConfig{
		Env:         yml.Deployment.Env,
		Wait:        false,
		ExposePorts: openPortsLocally,
		IsRootImage: true,
	}, &routableLinks, &liveContainerIDs)
	if err != nil {
		fail(ctx, err)
	}
	if len(yml.Deployment.Health) > 0 {
		_, err := utils.StartContainerByEazyYml(ctx, yml, models.GetLatestIntegrationImageName(yml), models.DockerConfig{
			Command:     yml.Deployment.Health,
			Wait:        true,
			ExposePorts: false,
		}, &routableLinks, &liveContainerIDs)
		if err != nil {
			fail(ctx, err)
		}
	}
}

func success(ctx context.Context) {
	cleanUp(ctx, 0, nil)
}

func fail(ctx context.Context, err error) {
	cleanUp(ctx, 1, err)
}

func cleanUp(ctx context.Context, exitCode int, err error) {
	log.Println("cleaning up running containers...")
	term.RestoreTerminal(os.Stdout.Fd(), oldStateOut)
	for _, id := range liveContainerIDs {
		err := utils.KillContainer(ctx, id)
		if err == nil {
			log.Println("container successfully shutdown: " + id)
		}
	}
	if exitCode == 0 {
		log.Println("Succeeded!")
	} else {
		log.Println(err)
		log.Println("CI Failed!")
	}
	os.Exit(exitCode)
}
