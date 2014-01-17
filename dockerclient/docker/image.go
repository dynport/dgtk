package docker

import (
	"fmt"
	"time"
)

// Basic image information as returned by docker.
type Image struct {
	Repository  string   `json:"Repository"`
	Tags        []string `json:"RepoTags"`
	Id          string   `json:"Id"`
	created     int
	Size        int
	VirtualSize int
}

// Image details as returned by docker.
type ImageDetails struct {
	Id              string           `json:"id"`
	Parent          string           `json:"parent"`
	Created         string           `json:"created"`
	Container       string           `json:"container"`
	Size            int              `json:"size"`
	Architecture    string           `json:"architecture"`
	DockerVersion   string           `json:"docker_version"`
	ContainerConfig *ContainerConfig `json:"container_config"`
	Config          *ContainerConfig `json:"config"`
}

// Image history entries as returned by docker.
type ImageHistory []struct {
	Id        string
	Tags      []string
	created   int
	CreatedBy string
}

func (self *Image) CreatedAt() time.Time {
	return time.Unix(int64(self.created), 0)
}

func (image *ImageDetails) CreatedAt() (time.Time, error) {
	t, e := time.Parse("2006-01-02T15:04:05.999999999Z", image.Created)
	if e != nil {
		t, e = time.Parse("2006-01-02T15:04:05.999999999-07:00", image.Created)
	}
	return t, e
}

func (img *Image) String() string {
	return fmt.Sprintf("%s:%s (%s)", img.Repository, img.Tags, img.CreatedAt().UTC())
}
