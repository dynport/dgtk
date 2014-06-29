package github

import (
	"fmt"
	"net/http"
	"strconv"
)

type LoadIssue struct {
	Repo   string
	Number int
}

func issueUrl(repo string, id int) string {
	return apiRoot + "/repos/" + repo + "/issues/" + strconv.Itoa(id)
}

func (i *LoadIssue) Execute(client *Client) (*Issue, error) {
	if i.Repo == "" {
		return nil, fmt.Errorf("Repo is not set")
	}
	req, e := http.NewRequest("GET", issueUrl(i.Repo, i.Number), nil)
	if e != nil {
		return nil, e
	}
	issue := &Issue{}
	e = client.loadRequest(req, issue)
	return issue, e
}
