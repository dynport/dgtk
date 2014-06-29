package main

import (
	"fmt"

	"github.com/dynport/dgtk/github"
)

type issueClose struct {
	Number int    `cli:"arg required"`
	Label  string `cli:"opt --label"`
}

func (r *issueClose) Run() error {
	repo, e := githubRepo()
	if e != nil {
		return e
	}
	if repo == "" {
		return fmt.Errorf("unable to get github repo from current path")
	}

	c, e := client()
	if e != nil {
		return e
	}
	issue, e := loadIssue(c, r.Number)
	if e != nil {
		return e
	}
	ci := &github.CreateIssue{State: "closed", Number: r.Number, Repo: repo}

	logger.Printf("closing issue %d", r.Number)
	if r.Label != "" {
		ci.Labels, e = addLabel(issue, r.Label)
		if e != nil {
			return e
		}
		logger.Printf("also assigning labels %q", ci.Labels)
	}
	issue, e = ci.Update(c)
	if e != nil {
		return e
	}
	return nil
}
