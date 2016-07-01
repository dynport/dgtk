package main

import (
	"fmt"
	"sort"

	"github.com/dynport/gocli"
	"github.com/google/go-github/github"
)

type reposList struct {
	Orga string `cli:"opt --orga required"`
}

func (r *reposList) Run() error {
	client, err := newClient()
	if err != nil {
		return err
	}
	var list []*github.Repository
	opts := &github.RepositoryListByOrgOptions{}
	opts.PerPage = 100

	for {
		current, headers, err := client.Repositories.ListByOrg(r.Orga, opts)
		if err != nil {
			return err
		}
		list = append(list, current...)
		if headers.NextPage == 0 {
			break
		}
		opts.Page = headers.NextPage
	}
	sort.Sort(repos(list))
	t := gocli.NewTable()
	for _, r := range list {
		if r.Name == nil {
			continue
		}
		t.Add(r.ID, colorizeName(r), r.StargazersCount, r.OpenIssuesCount, r.Private)
	}
	fmt.Println(t)
	return nil
}

type repos []*github.Repository

func (list repos) Len() int {
	return len(list)
}

func colorizeName(r *github.Repository) string {
	if r.Private != nil && *r.Private {
		return gocli.Red(*r.Name)
	}
	return gocli.Green(*r.Name)
}

func (list repos) Swap(a, b int) {
	list[a], list[b] = list[b], list[a]
}

func (list repos) Less(a, b int) bool {
	return p2s(list[a].Name) < p2s(list[b].Name)
}
