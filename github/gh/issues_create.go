package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type issuesCreate struct {
	Labels string `cli:"opt --labels"`
}

func (r *issuesCreate) Run() error {
	repo, e := githubRepo()
	if e != nil {
		return e
	}
	if repo == "" {
		return fmt.Errorf("unable to get github repo from current path")
	}
	logger.Printf("creating issues for %q", repo)
	scanner := bufio.NewScanner(os.Stdin)
	ci := &CreateIssue{}
	if len(r.Labels) > 0 {
		ci.Labels = strings.Split(r.Labels, ",")
	}
	if len(r.Labels) > 0 {
		fmt.Printf("Labels: %q\n", r.Labels)
	}
	fmt.Printf("Title: ")
	i := 0
	lines := []string{}
	for scanner.Scan() {
		if i == 0 {
			ci.Title = scanner.Text()
			fmt.Println(strings.Repeat("-", 100))
			fmt.Println("Body (send with ctrl+d):")
		} else {
			lines = append(lines, scanner.Text())
		}
		i++
	}
	e = scanner.Err()
	if e != nil {
		return e
	}
	ci.Body = strings.Join(lines, "\n")
	logger.Printf("finished scanning %d lines", len(lines))
	b, e := json.Marshal(ci)
	if e != nil {
		return e
	}

	dbg.Printf("postung to url %q")
	u := urlRoot + "/repos/" + repo + "/issues"
	rsp, e := authenticatedRequest("POST", u, bytes.NewReader(b))
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
	logger.Printf("created issue #%d", issue.Number)
	return nil
}
