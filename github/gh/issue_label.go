package main

import (
	"fmt"

	"github.com/dynport/dgtk/github"
)

type issueLabel struct {
	Number int    `cli:"arg required"`
	Label  string `cli:"arg required"`
}

func (r *issueLabel) Run() error {
	repo, e := githubRepo()
	if e != nil {
		return e
	}
	c, e := client()
	if e != nil {
		return e
	}
	issue, e := (&github.LoadIssue{Repo: repo, Number: r.Number}).Execute(c)
	if e != nil {
		return e
	}
	labels, e := addLabel(issue, r.Label)
	if e != nil {
		return e
	}

	ci := &github.CreateIssue{
		Repo:   repo,
		Number: r.Number,
		Labels: labels,
	}

	issue, e = ci.Update(c)
	if e != nil {
		return e
	}
	logger.Printf("labeled issue %d with label %q", r.Number, r.Label)
	return nil
}

func addLabel(issue *github.Issue, label string) ([]string, error) {
	labels := []string{}
	for _, l := range issue.Labels {
		if l.Name == label {
			return nil, fmt.Errorf("issue already labeled with label %q")
		}
		labels = append(labels, l.Name)
	}
	return append(labels, label), nil
}
