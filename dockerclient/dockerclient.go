// A docker client library.
package dockerclient

import "net/http"

type DockerHost struct {
	Address      string
	*http.Client // http client used to send requests to the host.
}

func New(address string, client *http.Client) *DockerHost {
	dh := &DockerHost{Address: address, Client: client}
	if dh.Client == nil {
		dh.Client = http.DefaultClient
	}
	return dh
}

const FAKE_AUTH = `
{
	"Auth": "fakeauth",
	"Email": "fake@email.xx"
}
`
