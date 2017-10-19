package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/dynport/gocli"
)

type teamsShow struct {
	ID int `cli:"arg required"`
}

func (r *teamsShow) Run() error {
	client, err := newClient()
	if err != nil {
		return err
	}
	team, _, err := client.Organizations.GetTeam(context.TODO(), r.ID)
	if err != nil {
		return err
	}
	// TODO: paginate
	repos, _, err := client.Organizations.ListTeamRepos(context.TODO(), r.ID, nil)
	if err != nil {
		return err
	}
	t := gocli.NewTable()
	t.Add("ID", team.ID)
	t.Add("Name", team.Name)
	fmt.Println(t)

	// TODO paginate
	members, _, err := client.Organizations.ListTeamMembers(context.TODO(), r.ID, nil)
	if err != nil {
		return err
	}

	t = gocli.NewTable()
	fmt.Println("\nMembers")
	t.Header("id", "login")
	for _, m := range members {
		t.Add(m.ID, m.Login)
	}
	fmt.Println(t)

	t = gocli.NewTable()
	fmt.Println("\nRepos")
	t.Header("id", "name", "permissions")
	for _, repo := range repos {
		t.Add(repo.ID, colorizeName(repo), formatPermissions(repo.Permissions))
	}
	fmt.Println(t)
	return nil
}

func formatPermissions(in *map[string]bool) string {
	out := []string{}
	if in == nil {
		return ""
	}
	for k, v := range *in {
		out = append(out, fmt.Sprintf("%s=%t", k, v))
	}
	sort.Strings(out)
	return strings.Join(out, ", ")
}
