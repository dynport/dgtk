package main

import "strconv"

type issueOpen struct {
	Number int `cli:"arg required"`
}

func (r *issueOpen) Run() error {
	return openGithubUrl("/issues/" + strconv.Itoa(r.Number))
}
