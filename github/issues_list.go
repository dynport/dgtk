package github

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ListIssues struct {
	Repo      string // orga/repo
	Milestone int
	State     string
	Assignee  string
	Creator   string
	Mentioned string
	Labels    []string
	Sort      string
	Direction string
	Since     time.Time
}

type values map[string]string

func (a *ListIssues) Execute(client *Client) ([]*Issue, error) {
	values := values{
		"state":     a.State,
		"assignee":  a.Assignee,
		"creator":   a.Creator,
		"mentioned": a.Mentioned,
		"sort":      a.Sort,
		"direction": a.Direction,
	}
	if a.Milestone > 0 {
		values["milestone"] = strconv.Itoa(a.Milestone)
	}
	if len(a.Labels) > 0 {
		values["labels"] = strings.Join(a.Labels, ",")
	}
	if !a.Since.IsZero() {
		values["since"] = a.Since.UTC().Format("2006-01-02T15:04:05Z")
	}

	v := url.Values{}
	for k, kv := range values {
		if kv != "" {
			v.Set(k, kv)

		}
	}
	path := "/issues"
	if a.Repo != "" {
		path = "/repos/" + a.Repo + "/issues"
	}
	if len(v) > 0 {
		path += "?" + v.Encode()
	}
	issues := []*Issue{}
	req, e := http.NewRequest("GET", ApiRoot+path, nil)
	if e != nil {
		return nil, e
	}
	e = client.loadRequest(req, &issues)
	return issues, e
}
