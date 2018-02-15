package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
)

type teamsAuthorize struct {
	ID         int64  `cli:"arg required"`
	Owner      string `cli:"arg required"`
	Repo       string `cli:"arg required"`
	Permission string `cli:"arg required"`
}

func (r *teamsAuthorize) Run() error {
	cl, err := newClient()
	if err != nil {
		return err
	}
	_, err = cl.Organizations.AddTeamRepo(context.TODO(), r.ID, r.Owner, r.Repo, &github.OrganizationAddTeamRepoOptions{Permission: r.Permission})
	if err != nil {
		return fmt.Errorf("authorizing team %d for repo %s/%s: %s", r.ID, r.Owner, r.Repo, err)
	}
	return nil
}
