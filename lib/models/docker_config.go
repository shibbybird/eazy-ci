package models

import "github.com/docker/docker/api/types/mount"

// DockerConfig meant for docker util commands
type DockerConfig struct {
	Env           []string
	Dockerfile    string
	Command       []string
	Mounts        []mount.Mount
	Wait          bool
	IsHostNetwork bool
	ExposePorts   bool
	Attach        bool
	IsRootImage   bool
	WorkingDir    string
}
