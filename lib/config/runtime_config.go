package config

import "github.com/docker/docker/api/types/mount"

// RuntimeConfig base config for all runtimes
type RuntimeConfig struct {
	User          string
	Env           []string
	Command       []string
	Mounts        []mount.Mount
	Wait          bool
	ExposePorts   bool
	Attach        bool
	IsRootImage   bool
	WorkingDir    string
	Dockerfile    string
	SkipImagePull bool
}
