package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
)

func New(token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("token must be provided")
	}
	return &Client{Client: NewHttpClient(token)}, nil
}

func NewHttpClient(token string) *http.Client {
	cl := &http.Client{}
	if token != "" {
		cl.Transport = &auth{token: token}
	}
	return cl
}

type Client struct {
	*http.Client
}

func (client *Client) loadRequest(req *http.Request, i interface{}) error {
	dbg.Printf("requesting %s with url %s", req.Method, req.URL.String())
	rsp, e := client.Do(req)
	if e != nil {
		return e
	}
	b, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return e
	}
	if rsp.Status[0] != '2' {
		return fmt.Errorf("expected status 2xx, got %s. body=%s", rsp.Status, string(b))
	}

	return json.Unmarshal(b, i)
}

type auth struct {
	token           string
	cachedTransport http.RoundTripper
}

func (auth *auth) transport() http.RoundTripper {
	if auth.cachedTransport == nil {
		auth.cachedTransport = http.DefaultTransport
	}
	return auth.cachedTransport
}

func (auth *auth) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", "token "+auth.token)
	return auth.transport().RoundTrip(r)
}

func githubToken() (token string, e error) {
	return readGitConfig("github.token")
}

func readGitConfig(name string) (string, error) {
	raw, e := exec.Command("git", "config", "--get", name).CombinedOutput()
	if e != nil {
		return "", fmt.Errorf("unable to get git config %q: %s", name, string(raw))
	}
	config := strings.TrimSpace(string(raw))
	if config == "" {
		return "", fmt.Errorf("no github config defined for %q", name)
	}
	return config, nil
}
