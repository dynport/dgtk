package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
)

type issueTag struct {
	Number int    `cli:"arg required"`
	Label  string `cli:"arg required"`
}

func (r *issueTag) Run() error {
	repo, e := githubRepo()
	if e != nil {
		return e
	}
	if repo == "" {
		return fmt.Errorf("unable to get github repo from current path")
	}
	u := urlRoot + "/repos/" + repo + "/issues/" + strconv.Itoa(r.Number)
	rsp, e := authenticatedRequest("GET", u, nil)
	if e != nil {
		return e
	}
	b, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return e
	}
	if rsp.Status[0] != '2' {
		return fmt.Errorf("expected status 2xx, got %s: %s", rsp.Status, string(b))
	}
	issue := &Issue{}
	e = json.Unmarshal(b, issue)
	if e != nil {
		return e
	}

	labels, e := addLabel(issue, r.Label)
	if e != nil {
		return e
	}

	ci := &CreateIssue{
		Labels: labels,
	}

	b, e = json.Marshal(ci)
	if e != nil {
		return e
	}
	rsp, e = authenticatedRequest("PATCH", u, bytes.NewReader(b))
	if e != nil {
		return e
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		return fmt.Errorf("expected status 2xx, got %s: %s", rsp.Status, string(b))
	}
	logger.Printf("tagged issue %d with label %q", r.Number, r.Label)
	return nil
}

func addLabel(issue *Issue, label string) ([]string, error) {
	labels := []string{}
	for _, l := range issue.Labels {
		if l.Name == label {
			return nil, fmt.Errorf("issue already tagged with label %q")
		}
		labels = append(labels, l.Name)
	}
	return append(labels, label), nil
}
