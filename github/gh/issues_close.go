package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
)

type issueClose struct {
	Ids []int `cli:"arg required"`
}

func (r *issueClose) Run() error {
	repo, e := githubRepo()
	if e != nil {
		return e
	}
	if repo == "" {
		return fmt.Errorf("unable to get github repo from current path")
	}
	for _, id := range r.Ids {
		logger.Printf("closing issue %d", id)
		u := urlRoot + "/repos/" + repo + "/issues/" + strconv.Itoa(id)
		issue := &Issue{State: "closed"}
		b, e := json.Marshal(issue)
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
		issue = &Issue{}
		e = json.Unmarshal(b, issue)
		if e != nil {
			return e
		}
	}
	return nil
}
