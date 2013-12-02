// A docker client library.
package dockerclient

import (
	"fmt"
	"github.com/dynport/gologger"
	"github.com/dynport/gossh"
	"net/http"
)

var logger = gologger.NewFromEnv()

type DockerHost struct {
	host       string
	port       int
	Registry   string       // Registry to use with this docker host.
	httpClient *http.Client // http client used to send requests to the host.
}

// Create a new connection to a docker host reachable at the given host and port.
func New(host string, port int) *DockerHost {
	return &DockerHost{host: host, port: port, httpClient: &http.Client{}}
}

// Create a new connection to a docker host using a SSH tunnel at the given user and host. This is useful as making the
// docker API public isn't recommended (allows to do stuff you don't want to allow publicly). A SSH connection will be
// set up and utilized for all communication to the docker API.
func NewViaTunnel(host, user string) (*DockerHost, error) {
	sc := gossh.New(host, user)
	hc, e := gossh.NewHttpClient(sc)
	if e != nil {
		return nil, e
	}
	return &DockerHost{httpClient: hc}, nil
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
