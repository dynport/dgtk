package main

import "github.com/dynport/dgtk/github"

func truncate(s string, l int, dots bool) string {
	if len(s) > l {
		if l > 6 && dots {
			return s[0:l-3] + "..."
		}
		return s[0:l]
	}
	return s
}

func loadIssue(client *github.Client, id int) (*github.Issue, error) {
	repo, e := githubRepo()
	if e != nil {
		return nil, e
	}
	a := github.LoadIssue{Repo: repo, Number: id}
	return a.Execute(client)
}
