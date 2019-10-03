package models

// DockerConfig meant for docker util commands
type DockerConfig struct {
	Env           []string
	Dockerfile    string
	Command       []string
	Wait          bool
	IsHostNetwork bool
	ExposePorts   bool
	Attach        bool
	IsRootImage   bool
}
