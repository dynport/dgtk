package dockerclient

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/dynport/dgtk/dockerclient/docker"
	"io"
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

// Attach to the given container with the given writer.
func (dh *DockerHost) AttachContainer(containerId string, w io.Writer) (e error) {
	rsp, e := dh.post(dh.url() + "/containers/" + containerId + "/attach?logs=1&stream=1&stdout=1&stderr=1")
	if e != nil {
		return e
	}
	defer rsp.Body.Close()

	if w != nil {
		scanner := bufio.NewScanner(rsp.Body)
		for scanner.Scan() {
			bytes := scanner.Bytes()
			i := 0
			for i < len(bytes) {
				n, e := w.Write(bytes[i:])
				if e != nil {
					return e
				}
				i += n
			}
		}
	}
	return nil
}
