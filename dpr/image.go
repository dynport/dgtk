package main

import (
	"time"
)

type Image struct {
	Id              string           `json:"id"`
	Tag             string           `json:"tag"`
	Comment         string           `json:"comment"`
	Created         time.Time        `json:"created"`
	ContainerConfig *ContainerConfig `json:"container_config"`
	DockerVersion   string           `json:"docker_version"`
	Parent          string           `json:"parent"`
}
