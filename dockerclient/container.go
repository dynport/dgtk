package dockerclient

import (
	"errors"
	"fmt"
	"github.com/dynport/dgtk/dockerclient/docker"
	"io"
	"net/url"
)

// Get a list of all ontainers available on the host.
func (dh *DockerHost) Containers() (containers []*docker.Container, e error) {
	e = dh.getJSON(dh.url()+"/containers/json", &containers)
	return containers, e
}

// Get the information for the container with the given id.
func (dh *DockerHost) Container(containerId string) (containerInfo *docker.ContainerInfo, e error) {
	containerInfo = &docker.ContainerInfo{}
	e = dh.getJSON(dh.url()+"/containers/"+containerId+"/json", containerInfo)
	return containerInfo, e
}

// For the given image name and the given container configuration, create a container. If the image name deosn't contain
// a tag "latest" is used by default.
func (dh *DockerHost) CreateContainer(imageName string, options *docker.ContainerConfig) (containerId string, e error) {
	// Verify image available on host.
	_, e = dh.ImageHistory(imageName)
	if e != nil && e.Error() == "resource not found" {
		if e = dh.PullImage(imageName); e != nil {
			return "", e
		}
	}

	if options == nil {
		options = &docker.ContainerConfig{}
	}
	options.Image = imageName

	container := &docker.Container{}
	content, _, e := dh.postJSON(dh.url()+"/containers/create", options, container)
	if e != nil {
		return "", fmt.Errorf("failed creating container (%s): %s", e.Error(), content)
	}
	return container.Id, e
}

// Start the container with the given identifier. The hostConfig can safely be set to nil to use the defaults.
func (dh *DockerHost) StartContainer(containerId string, hostConfig *docker.HostConfig) (e error) {
	if hostConfig == nil {
		hostConfig = &docker.HostConfig{}
	}
	dh.Logger.Infof("starting container with binds %+v", hostConfig)
	body, rsp, e := dh.postJSON(dh.url()+"/containers/"+containerId+"/start", hostConfig, nil)
	if e != nil {
		return
	}
	if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		e = errors.New(fmt.Sprintf("error starting container %s: status=%s, response=%s", containerId, rsp.StatusCode, string(body)))
	} else {
		dh.Logger.Infof("started container %s", containerId)
	}
	return
}

// Kill the container with the given identifier.
func (dh *DockerHost) StopContainer(containerId string) (e error) {
	rsp, e := dh.post(dh.url() + "/containers/" + containerId + "/kill")
	defer rsp.Body.Close()
	return e
}

type AttachOptions struct {
	Logs   bool
	Stream bool
	Stdout bool
	Stderr bool
}

func (opts *AttachOptions) Encode() string {
	v := url.Values{}
	if opts.Logs {
		v.Add("logs", "1")
	}
	if opts.Stream {
		v.Add("stream", "1")
	}
	if opts.Stdout {
		v.Add("stdout", "1")
	}
	if opts.Stderr {
		v.Add("stderr", "1")
	}
	if len(v) > 0 {
		return "?" + v.Encode()
	}
	return ""
}

// Attach to the given container with the given writer.
func (dh *DockerHost) AttachContainer(containerId string, w io.Writer, opts *AttachOptions) (e error) {
	if opts == nil {
		opts = &AttachOptions{}
	}
	rsp, e := dh.post(dh.url() + "/containers/" + containerId + "/attach" + opts.Encode())
	if e != nil {
		return e
	}
	defer rsp.Body.Close()

	_, e = io.Copy(w, rsp.Body)
	return e
}
