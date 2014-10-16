package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
)

type ListPullRequests struct {
	Org       string
	Repo      string
	State     string
	Sort      string
	Direction string
}

func (a *ListPullRequests) Execute(client *Client) ([]*PullRequest, error) {
	values := values{
		"state":     a.State,
		"sort":      a.Sort,
		"direction": a.Direction,
	}

	v := url.Values{}
	for k, kv := range values {
		if kv != "" {
			v.Set(k, kv)

		}
	}
	theUrl := ApiRoot + "/repos/" + a.Org + "/" + a.Repo + "/pulls"
	if len(v) > 0 {
		theUrl += "?" + v.Encode()
	}
	rsp, e := client.Get(theUrl)

	if e != nil {
		return nil, e
	}
	defer rsp.Body.Close()
	b, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return nil, e
	}
	if rsp.Status[0] != '2' {
		return nil, fmt.Errorf("expected status 2xx, got %s: %s", rsp.Status, string(b))
	}
	prs := []*PullRequest{}
	e = json.Unmarshal(b, &prs)
	return prs, e
}
