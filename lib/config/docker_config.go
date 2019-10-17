package config

import "github.com/docker/docker/api/types/mount"

// DockerConfig meant for docker util commands
type DockerConfig struct {
	User          string
	Env           []string
	Dockerfile    string
	Command       []string
	Mounts        []mount.Mount
	Wait          bool
	ExposePorts   bool
	Attach        bool
	IsRootImage   bool
	WorkingDir    string
	SkipImagePull bool
}
