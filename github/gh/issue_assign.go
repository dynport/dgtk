package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
)

type issueAssign struct {
	Number   int    `cli:"arg required"`
	Assignee string `cli:"arg required"`
}

func (r *issueAssign) Run() error {
	repo, e := githubRepo()
	if e != nil {
		return e
	}
	if repo == "" {
		return fmt.Errorf("could not find repo from current directory")
	}
	ci := &CreateIssue{
		Assignee: r.Assignee,
	}

	u := urlRoot + "/repos/" + repo + "/issues/" + strconv.Itoa(r.Number)
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
	issue := &Issue{}
	e = json.Unmarshal(b, issue)
	if e != nil {
		return e
	}
	if issue.Assignee != nil {
		logger.Printf("assigned issue #%d to %q", issue.Number, issue.Assignee.Login)
	}
	return nil
}
