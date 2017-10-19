package main

import (
	"context"
	"os"

	"github.com/google/go-github/github"
)

type reposCreate struct {
	Name     string `cli:"arg required"`
	Public   bool   `cli:"opt --public"`
	Orga     string `cli:"opt --orga"`
	Teams    []int  `cli:"opt --teams"`
	ReadOnly bool   `cli:"opt --read-only desc='Add teams with read only'"`
	Clone    bool   `cli:"opt --clone"`
}

func (r *reposCreate) Run() error {
	cl, err := newClient()
	if err != nil {
		return err
	}
	repo := &github.Repository{Name: s2p(r.Name), Private: b2p(!r.Public)}

	if _, _, err := cl.Repositories.Create(context.TODO(), r.Orga, repo); err != nil {
		return err
	}
	permission := "push"
	if r.ReadOnly {
		permission = "pull"
	}
	for _, t := range r.Teams {
		_, err = cl.Organizations.AddTeamRepo(context.TODO(), t, r.Orga, r.Name, &github.OrganizationAddTeamRepoOptions{Permission: permission})
		if err != nil {
			return err
		}
	}
	if r.Clone {
		repo := "git@github.com:" + r.Orga + "/" + r.Name + ".git"
		return cloneRepo(repo, os.ExpandEnv("$GOPATH/src/github.com/"+r.Orga+"/"+repo))
	}
	return nil
}
