package main

import (
	"context"
	"fmt"

	"github.com/dynport/gocli"
)

type reposShow struct {
	Owner string `cli:"arg required"`
	Name  string `cli:"arg required"`
}

func (r *reposShow) Run() error {
	cl, err := newClient()
	if err != nil {
		return err
	}
	res, _, err := cl.Repositories.Get(context.TODO(), r.Owner, r.Name)
	if err != nil {
		return err
	}
	t := gocli.NewTable()
	t.Add("ID", res.ID)
	t.Add("Name", res.Name)
	t.Add("Issues", res.OpenIssuesCount)
	colabs, _, err := cl.Repositories.ListCollaborators(context.TODO(), r.Owner, r.Name, nil)
	if err != nil {
		return err
	}
	for _, c := range colabs {
		t.Add("Colaborator", c.Login, c.Email)
	}
	fmt.Println(t)

	return nil
}
