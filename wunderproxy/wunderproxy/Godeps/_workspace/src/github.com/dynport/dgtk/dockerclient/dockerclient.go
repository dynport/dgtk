// A docker client library.
package dockerclient

import (
	"fmt"
	"net/http"

	"github.com/dynport/dgtk/wunderproxy/wunderproxy/Godeps/_workspace/src/github.com/dynport/dgtk/version"
	"github.com/dynport/dgtk/wunderproxy/wunderproxy/Godeps/_workspace/src/github.com/dynport/gossh"
)

type DockerHost struct {
	host       string
	port       int
	Registry   string       // Registry to use with this docker host.
	httpClient *http.Client // http client used to send requests to the host.

	cachedServerVersion *Version
}

func newDockerHost(host string, port int, hc *http.Client) (*DockerHost, error) {
	dh := &DockerHost{host: host, port: port, httpClient: hc}
	dockerVersion, e := dh.ServerVersion()
	if e != nil {
		return nil, e
	}
	v, e := version.Parse(dockerVersion.Version)
	if e != nil {
		return nil, e
	}
	if v.Less(&version.Version{Major: 1, Minor: 2}) {
		return nil, fmt.Errorf("at least docker version 1.2 required (found %s)", dockerVersion.Version)
	}
	return dh, nil
}

// Create a new connection to a docker host reachable at the given host and port.
func New(host string, port int) (*DockerHost, error) {
	return newDockerHost(host, port, &http.Client{})
}

// Create a new connection to a docker host using a SSH tunnel at the given user and host. This is useful as making the
// docker API public isn't recommended (allows to do stuff you don't want to allow publicly). A SSH connection will be
// set up and utilized for all communication to the docker API.
func NewViaTunnel(host, user, password string) (*DockerHost, error) {
	sc := gossh.New(host, user)
	sc.SetPassword(password)

	hc, e := gossh.NewHttpClient(sc)
	if e != nil {
		return nil, e
	}
	return newDockerHost(host, 4243, hc)
}

func (dh *DockerHost) url() string {
	host := dh.host
	if host == "" {
		host = "127.0.0.1"
	}

	port := dh.port
	if port == 0 {
		port = 4243
	}

	return fmt.Sprintf("http://%s:%d", host, port)
}

const FAKE_AUTH = `
{
	"Auth": "fakeauth",
	"Email": "fake@email.xx"
}
`
