package models

import (
	"io/ioutil"
	"testing"
)

func TestMicroYml(t *testing.T) {
	data, err := ioutil.ReadFile("./testdata/test-micro.yml")

	if err != nil {
		t.Error(err)
	}

	micro, err := MicroYmlUnmarshal(data)

	if err != nil {
		t.Error(err)
	}

	if micro.MicroVersion != "1.0" ||
		micro.Releases[0] != "v1.0.0" ||
		micro.Releases[1] != "v1.1.0" ||
		micro.Image != "test/image" ||
		micro.Deployment.Health[0] != "/bin/sh" ||
		micro.Deployment.Health[1] != "-c" ||
		micro.Deployment.Health[2] != "while ! curl http://host.docker.internal:8080/health; do sleep 1; done;" ||
		micro.Deployment.Ports[0] != 9000 ||
		micro.Integration.Bootstrap[0] != "/bin/sh" ||
		micro.Integration.Bootstrap[1] != "-c" ||
		micro.Integration.Bootstrap[2] != "liquibase" ||
		micro.Integration.RunTest[0] != "/bin/sh" ||
		micro.Integration.RunTest[1] != "-c" ||
		micro.Integration.RunTest[2] != "npm test" ||
		micro.Integration.Dependencies[0] != "github.com/shibbybird/test-api" ||
		micro.Integration.Dependencies[1] != "github.com/shibbybird/test-service" ||
		micro.Integration.PeerDependencies[0] != "github.com/shibbybird/cassandra-db" {
		t.Error(micro)
	}

}
