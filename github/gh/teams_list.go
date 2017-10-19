package main

import (
	"context"
	"fmt"

	"github.com/dynport/gocli"
)

type teamsList struct {
	Orga string `cli:"arg required"`
}

func (r *teamsList) Run() error {
	client, err := newClient()
	if err != nil {
		return err
	}
	// TODO: paginate
	res, _, err := client.Organizations.ListTeams(context.TODO(), r.Orga, nil)
	if err != nil {
		return err
	}
	t := gocli.NewTable()
	t.Add("id", "name")
	for _, team := range res {
		t.Add(team.ID, team.Name)
	}
	fmt.Println(t)
	return nil
}
