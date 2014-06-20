package main

import (
	"fmt"
	"strings"

	"github.com/dynport/gocli"
)

type issuesList struct {
	All bool `cli:"opt --all"`
}

func (r *issuesList) Run() error {
	var e error
	repo := ""
	if !r.All {
		repo, e = githubRepo()
		if e != nil {
			logger.Printf("ERROR=%q", e)
		}
	}
	issues, e := loadIssues(repo)
	if e != nil {
		return e
	}
	if len(issues) == 0 {
		fmt.Println("no issues found")
		return nil
	}
	t := gocli.NewTable()
	for _, i := range issues {
		orga, issueRepo, e := i.Repo()
		if e != nil {
			return e
		}
		labels := []string{}
		for _, l := range i.Labels {
			labels = append(labels, l.Name)
		}
		parts := []interface{}{i.Number}
		if repo == "" {
			parts = append(parts, orga+"/"+issueRepo)
		}
		assignee := ""
		if i.Assignee != nil {
			assignee = i.Assignee.Login
		}
		parts = append(parts, truncate(i.Title, 48, true), truncate(i.CreatedAt, 16, false), assignee, i.State, strings.Join(labels, ","))
		t.Add(parts...)
	}
	fmt.Println(t)
	return nil
}
