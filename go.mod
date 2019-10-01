module github.com/shibbybird/eazy-ci

go 1.13

require (
	github.com/Microsoft/go-winio v0.4.14
	github.com/Microsoft/hcsshim v0.8.6
	github.com/containerd/containerd v1.2.9
	github.com/containerd/continuity v0.0.0-20190827140505-75bee3e2ccb6
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0
	github.com/emirpasic/gods v1.12.0
	github.com/gogo/protobuf v1.3.0
	github.com/golang/protobuf v1.3.2
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99
	github.com/kevinburke/ssh_config v0.0.0-20190724205821-6cfae18c12b8
	github.com/konsorten/go-windows-terminal-sequences v1.0.2
	github.com/mitchellh/go-homedir v1.1.0
	github.com/moby/moby v1.4.2-0.20190924232817-ef89d70aed01
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v0.1.1
	github.com/pkg/errors v0.8.1
	github.com/sergi/go-diff v1.0.0
	github.com/sirupsen/logrus v1.4.2
	github.com/src-d/gcfg v1.4.0
	github.com/xanzy/ssh-agent v0.2.1
	golang.org/x/crypto v0.0.0-20190923035154-9ee001bba392
	golang.org/x/net v0.0.0-20190923162816-aa69164e4478
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20190924154521-2837fb4f24fe
	google.golang.org/genproto v0.0.0-20190916214212-f660b8655731
	google.golang.org/grpc v1.23.1
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/warnings.v0 v0.1.2
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/docker/docker ef89d70aed01d05adde6b9f3cdeba1c90d87bde8 => github.com/moby/moby v1.4.2-0.20190924232817-ef89d70aed01
