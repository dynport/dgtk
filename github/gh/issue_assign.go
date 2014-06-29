package main

import (
	"fmt"

	"github.com/dynport/dgtk/github"
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
	ci := &github.CreateIssue{
		Repo:     repo,
		Number:   r.Number,
		Assignee: r.Assignee,
	}

	c, e := client()
	if e != nil {
		return e
	}

	issue, e := ci.Update(c)

	if issue.Assignee != nil {
		logger.Printf("assigned issue #%d to %q", issue.Number, issue.Assignee.Login)
	}
	return nil
}
