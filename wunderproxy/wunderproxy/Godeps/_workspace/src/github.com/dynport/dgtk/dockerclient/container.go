package dockerclient

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dynport/dgtk/wunderproxy/wunderproxy/Godeps/_workspace/src/github.com/dynport/dgtk/dockerclient/docker"
)

type ListContainersOptions struct {
	All    bool
	Limit  int
	Since  string
	Before string
	Size   bool
}

func (opts *ListContainersOptions) Encode() string {
	values := url.Values{}
	if opts.All {
		values.Add("all", "true")
	}
	if opts.Limit > 0 {
		values.Add("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Since != "" {
		values.Add("since", opts.Since)
	}
	if opts.Before != "" {
		values.Add("before", opts.Before)
	}
	if opts.Size {
		values.Add("size", "true")
	}

	if len(values) > 0 {
		return values.Encode()
	}
	return ""
}

// Get a list of all ontainers available on the host.
func (dh *DockerHost) Containers() (containers []*docker.Container, e error) {
	return dh.ListContainers(nil)
}

func (dh *DockerHost) ListContainers(opts *ListContainersOptions) (containers []*docker.Container, e error) {
	u := dh.url() + "/containers/json"
	if opts != nil {
		if params := opts.Encode(); params != "" {
			u += "?" + params
		}
	}
	e = dh.getJSON(u, &containers)
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
func (dh *DockerHost) CreateContainer(options *docker.ContainerConfig, name string) (containerId string, e error) {
	imageId := options.Image

	// Verify image available on host.
	_, e = dh.ImageHistory(imageId)
	if e != nil && e.Error() == "resource not found" {
		if e = dh.PullImage(imageId); e != nil {
			return "", e
		}
	}

	imgDetails, err := dh.ImageDetails(imageId)
	if err != nil {
		return "", err
	}
	options.Env = append(options.Env, "DOCKER_IMAGE="+imgDetails.Id)

	container := &docker.Container{}
	u := dh.url() + "/containers/create"

	if name != "" {
		u += "?name=" + name
	}

	if _, e = dh.postJSON(u, options, container); e != nil {
		return "", fmt.Errorf("failed creating container: %s", e.Error())
	}
	return container.Id, e
}

// Start the container with the given identifier. The hostConfig can safely be set to nil to use the defaults.
func (dh *DockerHost) StartContainer(containerId string, hostConfig *docker.HostConfig) (e error) {
	if hostConfig == nil {
		hostConfig = &docker.HostConfig{}
	}
	_, e = dh.postJSON(dh.url()+"/containers/"+containerId+"/start", hostConfig, nil)
	return e
}

func (dh *DockerHost) RemoveContainer(containerId string) error {
	req, e := http.NewRequest("DELETE", dh.url()+"/containers/"+containerId, nil)
	if e != nil {
		return e
	}
	rsp, e := dh.httpClient.Do(req)
	if e != nil {
		return e
	}
	if rsp.Status[0] != '2' {
		return fmt.Errorf("expected status 2xx but got %s", rsp.Status)
	}
	return nil
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
	Stdout io.Writer
	Stderr io.Writer
}

func (opts *AttachOptions) Encode() string {
	v := url.Values{}
	if opts.Logs {
		v.Add("logs", "1")
	}
	if opts.Stream {
		v.Add("stream", "1")
	}
	if opts.Stdout != nil {
		v.Add("stdout", "1")
	}
	if opts.Stderr != nil {
		v.Add("stderr", "1")
	}
	if len(v) > 0 {
		return "?" + v.Encode()
	}
	return ""
}

func messageLength(header []byte) int {
	msgLength := int(header[7]) << 0
	msgLength += int(header[6]) << 8
	msgLength += int(header[5]) << 16
	msgLength += int(header[4]) << 24
	return msgLength
}

// See http://docs.docker.io/en/latest/api/docker_remote_api_v1.8/#attach-to-a-container for the stream protocol.
func handleMessages(r io.Reader, stdout io.Writer, stderr io.Writer) error {
	headerBuf := make([]byte, 8)
	for {
		n, e := r.Read(headerBuf)
		if e != nil {
			if e == io.EOF {
				return nil
			}
			return e
		}
		if n != 8 {
			return fmt.Errorf("failed reading; header to short")
		}

		msgLength := messageLength(headerBuf)
		msgBuf := make([]byte, msgLength) // buffer size taken from io.Copy
		n = 0
		for n < msgLength {
			i, e := r.Read(msgBuf[n:])
			if e != nil {
				return e
			}
			n += i
		}

		switch headerBuf[0] {
		case 0: // stdin
			if stdout != nil {
				_, _ = stdout.Write([]byte{'+'})
			}
		case 1: // stdout
			if stdout != nil {
				_, e := stdout.Write(msgBuf[:msgLength])
				if e != nil {
					return e
				}
			}
		case 2: // stderr
			if stderr != nil {
				_, e := stderr.Write(msgBuf[:msgLength])
				if e != nil {
					return e
				}
			}
		default:
			return fmt.Errorf("unknown stream source received")
		}
	}
}

// Attach to the given container with the given writer.
func (dh *DockerHost) AttachContainer(containerId string, opts *AttachOptions) (e error) {
	if opts == nil {
		opts = &AttachOptions{}
	}
	rsp, e := dh.post(dh.url() + "/containers/" + containerId + "/attach" + opts.Encode())
	if e != nil {
		return e
	}
	defer rsp.Body.Close()

	return handleMessages(rsp.Body, opts.Stdout, opts.Stderr)
}
