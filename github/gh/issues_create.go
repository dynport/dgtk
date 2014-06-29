package main

import (
	"bufio"

	"fmt"
	"os"
	"strings"

	"github.com/dynport/dgtk/github"
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
	ci := &github.CreateIssue{Repo: repo}
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

	c, e := client()
	if e != nil {
		return e
	}
	issue, e := ci.Create(c)
	if e != nil {
		return e
	}
	logger.Printf("created issue #%d", issue.Number)
	return nil
}
