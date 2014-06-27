package github

import (
	"fmt"
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
	return &http.Client{
		Transport: &auth{token: token},
	}
}

type Client struct {
	*http.Client
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
