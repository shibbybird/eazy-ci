package main

import (
	"context"
	"fmt"
)

var liveContainerIds = []string{}

func main() {
	ctx := context.Background()
	/*
		filePath := flag.String("f", "./eazy.yml", "The Eazy CI file ")
		isDev := flag.Bool("d", false, "Run dependencies and peer depedencies")
		isIntegration := flag.Bool("i", false, "Run dependencies, peer dependencies, and build/start Dockerfile")
		isHostMode := flag.Bool("h", false, "Sets docker to host mode")
		pemKeyPath := flag.String("k", "", "File path for ssh private key for github access")

		flag.Parse()

		data, err := ioutil.ReadFile(*filePath)
		if err != nil {

		}
	*/
	success(ctx)

}

func success(ctx context.Context) {
	cleanUp(0, ctx)
}

func fail(ctx context.Context) {
	cleanUp(1, ctx)
}

func cleanUp(exitCode int, ctx context.Context) {
	fmt.Println("Do Clean Up")
}
