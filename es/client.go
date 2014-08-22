package es

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Client struct {
	Address string
}

func (c *Client) Stats() (*Stats, error) {
	var s *Stats
	return s, c.load("/_stats", &s)
}

func (c *Client) load(path string, i interface{}) error {
	if c.Address == "" {
		return fmt.Errorf("Address must be set")
	}
	u := strings.TrimSuffix(c.Address, "/") + "/" + strings.TrimPrefix(path, "/")
	logger.Printf("sending req %s", u)

	rsp, e := http.Get(u)
	if e != nil {
		return nil
	}
	defer rsp.Body.Close()
	if e != nil {
		return e
	}
	if rsp.Status[0] != '2' {
		return fmt.Errorf("expected status 2xx, got %s", rsp.Status)
	}
	return json.NewDecoder(rsp.Body).Decode(&i)
}
