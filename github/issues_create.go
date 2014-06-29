package github

import (
	"bytes"
	"encoding/json"
	"fmt"

	"net/http"
)

type CreateIssue struct {
	Repo      string   `json:"-"`
	Number    int      `json:"-"`
	Title     string   `json:"title,omitempty"`
	Body      string   `json:"body,omitempty"`
	Assignee  string   `json:"assignee,omitempty"`
	Milestone int      `json:"milestone,omitempty"`
	State     string   `json:"state,omitempty"`
	Labels    []string `json:"labels,omitempty"`
}

func (c *CreateIssue) Create(client *Client) (*Issue, error) {
	if c.Repo == "" {
		return nil, fmt.Errorf("Repo must be provided on Create")
	}
	b, e := json.Marshal(c)
	if e != nil {
		return nil, e
	}
	req, e := http.NewRequest("POST", apiRoot+"/repos/"+c.Repo+"/issues", bytes.NewReader(b))
	if e != nil {
		return nil, e
	}
	i := &Issue{}
	e = client.loadRequest(req, i)
	return i, e

}

func (c *CreateIssue) Update(client *Client) (*Issue, error) {
	if c.Repo == "" {
		return nil, fmt.Errorf("Repo must be provided on Update")
	}
	if c.Number < 1 {
		return nil, fmt.Errorf("Number be be set on Update")
	}
	b, e := json.Marshal(c)
	if e != nil {
		return nil, e
	}
	req, e := http.NewRequest("PATCH", issueUrl(c.Repo, c.Number), bytes.NewReader(b))
	if e != nil {
		return nil, e
	}
	i := &Issue{}
	e = client.loadRequest(req, i)
	return i, e
}
