package main

import (
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/shibbybird/eazy-ci/lib/utils"

	"github.com/shibbybird/eazy-ci/lib/models"
)

var liveContainerIDs = []string{}

func main() {
	ctx := context.Background()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			cleanUp(ctx, 1, nil)
		}
	}()

	filePath := flag.String("f", "./eazy.yml", "The Eazy CI file ")
	isDev := flag.Bool("d", false, "Run dependencies and peer depedencies")
	isIntegration := flag.Bool("i", false, "Run dependencies, peer dependencies, and build/start Dockerfile")
	isHostMode := flag.Bool("h", false, "Sets docker to host mode")
	pemKeyPath := flag.String("k", "", "File path for ssh private key for github access")

	flag.Parse()

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

	log.Println(peerDependencies)

	for _, d := range peerDependencies {
		startUnit(ctx, d, *isHostMode)
	}

	for _, d := range dependencies {
		startUnit(ctx, d, *isHostMode)
	}

	if len(yml.Integration.Bootstrap) > 0 {
		containerID, err := utils.BuildAndRunContainer(ctx, "", yml, "Integration.Dockerfile", yml.Integration.Bootstrap, true, *isHostMode, false)
		if len(containerID) > 0 {
			liveContainerIDs = append(liveContainerIDs, containerID)
		}
		if err != nil {
			fail(ctx, err)
		}
	}

	if !*isDev {
		containerID, err := utils.BuildAndRunContainer(ctx, "", yml, "Dockerfile", []string{}, false, *isHostMode, true)
		if len(containerID) > 0 {
			liveContainerIDs = append(liveContainerIDs, containerID)
		}
		if err != nil {
			fail(ctx, err)
		}

		if len(yml.Deployment.Health) > 0 {
			containerID, err := utils.BuildAndRunContainer(ctx, "", yml, "Integration.Dockerfile", yml.Integration.Bootstrap, true, *isHostMode, false)
			if len(containerID) > 0 {
				liveContainerIDs = append(liveContainerIDs, containerID)
			}
			if err != nil {
				fail(ctx, err)
			}
		}
	}

	if *isDev || *isIntegration {
		log.Println("You are running in a Development Mode. Use ctrl-c to exit at anytime.")
		go forever()
		select {}
	} else {
		containerID, err := utils.BuildAndRunContainer(ctx, "", yml, "Integration.Dockerfile", yml.Integration.RunTest, true, *isHostMode, false)
		if len(containerID) > 0 {
			liveContainerIDs = append(liveContainerIDs, containerID)
		}
		if err != nil {
			fail(ctx, err)
		}
		success(ctx)
	}

	success(ctx)
}

func startUnit(ctx context.Context, yml models.EazyYml, isHostMode bool) {
	if len(yml.Integration.Bootstrap) > 0 {
		containerID, err := utils.StartContainerByEazyYml(ctx, yml, yml.Integration.Bootstrap, true, isHostMode, false, models.GetLatestIntegrationImageName(yml))
		if len(containerID) > 0 {
			liveContainerIDs = append(liveContainerIDs, containerID)
		}
		if err != nil {
			fail(ctx, err)
		}
	}
	containerID, err := utils.StartContainerByEazyYml(ctx, yml, yml.Integration.Bootstrap, false, isHostMode, true, "")
	if err != nil {
		fail(ctx, err)
	}
	liveContainerIDs = append(liveContainerIDs, containerID)
	if len(yml.Deployment.Health) > 0 {
		containerID, err := utils.StartContainerByEazyYml(ctx, yml, yml.Deployment.Health, true, isHostMode, false, models.GetLatestIntegrationImageName(yml))
		if len(containerID) > 0 {
			liveContainerIDs = append(liveContainerIDs, containerID)
		}
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
	log.Println("Do Clean Up!")
	for _, id := range liveContainerIDs {
		err := utils.KillContainer(ctx, id)
		if err != nil {
			log.Println("container already shutdown: " + id)
		}
	}
	log.Println(err)
	os.Exit(exitCode)
}

func forever() {
	for {
		time.Sleep(time.Second)
	}
}
