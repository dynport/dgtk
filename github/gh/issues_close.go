package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
)

type issueClose struct {
	Id  int    `cli:"arg required"`
	Tag string `cli:"opt --tag"`
}

func (r *issueClose) Run() error {
	repo, e := githubRepo()
	if e != nil {
		return e
	}
	if repo == "" {
		return fmt.Errorf("unable to get github repo from current path")
	}

	issue, e := loadIssue(r.Id)
	if e != nil {
		return e
	}

	u := urlRoot + "/repos/" + repo + "/issues/" + strconv.Itoa(r.Id)
	ci := &CreateIssue{State: "closed"}

	logger.Printf("closing issue %d", r.Id)
	if r.Tag != "" {
		ci.Labels, e = addLabel(issue, r.Tag)
		if e != nil {
			return e
		}
		logger.Printf("also assigning labels %q", ci.Labels)
	}
	b, e := json.Marshal(ci)
	if e != nil {
		return e
	}
	rsp, e := authenticatedRequest("PATCH", u, bytes.NewReader(b))
	if e != nil {
		return e
	}
	defer rsp.Body.Close()
	b, e = ioutil.ReadAll(rsp.Body)
	if e != nil {
		return e
	}
	if rsp.Status[0] != '2' {
		return fmt.Errorf("expected status 2xx, got %s: %s", rsp.Status, string(b))
	}
	return nil
}
