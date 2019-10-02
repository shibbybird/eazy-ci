package models

type DockerConfig struct {
	EazyVersion string
	Releases    []string
	Image       string
	Deployment  struct {
		Ports  []string
		Health []string
	}
	Integration struct {
		Bootstrap        []string
		RunTest          []string
		Dependencies     []string
		PeerDependencies []string
	}
}
