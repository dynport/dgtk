// A docker client library.
package dockerclient

import "net/http"

type Client struct {
	Address string
	Client  *http.Client // http client used to send requests to the host.
}

// Create a new connection to a docker host reachable at the given host and port.
func New(addr string) *Client {
	return &Client{Address: addr, Client: &http.Client{}}
}

const FAKE_AUTH = `
{
	"Auth": "fakeauth",
	"Email": "fake@email.xx"
}
`
