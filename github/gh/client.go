package main

import (
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func newClient() (*github.Client, error) {
	token, err := githubToken()
	if err != nil {
		return nil, err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	return github.NewClient(tc), nil
}
