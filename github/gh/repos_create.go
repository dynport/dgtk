package main

import "github.com/google/go-github/github"

type reposCreate struct {
	Name     string `cli:"arg required"`
	Public   bool   `cli:"opt --public"`
	Orga     string `cli:"opt --orga"`
	Teams    []int  `cli:"opt --teams"`
	ReadOnly bool   `cli:"opt --read-only desc='Add teams with read only'"`
}

func (r *reposCreate) Run() error {
	cl, err := newClient()
	if err != nil {
		return err
	}
	repo := &github.Repository{Name: s2p(r.Name), Private: b2p(!r.Public)}

	if _, _, err := cl.Repositories.Create(r.Orga, repo); err != nil {
		return err
	}
	permission := "push"
	if r.ReadOnly {
		permission = "pull"
	}
	for _, t := range r.Teams {
		_, err = cl.Organizations.AddTeamRepo(t, r.Orga, r.Name, &github.OrganizationAddTeamRepoOptions{Permission: permission})
		if err != nil {
			return err
		}
	}
	return nil
}
