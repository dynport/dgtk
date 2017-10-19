package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type issuesCommit struct {
	ID  int  `cli:"arg required"`
	Ref bool `cli:"opt --ref"`
	Fix bool `cli:"opt --fix"`
}

func (r *issuesCommit) Run() error {
	cl, err := newClient()
	if err != nil {
		return err
	}
	_ = cl
	raw, err := githubRepo()
	if err != nil {
		return err
	}
	parts := strings.SplitN(raw, "/", 2)
	if len(parts) != 2 || parts[1] == "" || parts[0] == "" {
		return fmt.Errorf("unable to extract owner and repo from %q", raw)
	}
	owner, repo := parts[0], parts[1]
	is, _, err := cl.Issues.Get(context.TODO(), owner, repo, r.ID)
	if err != nil {
		return err
	}
	verb := "closes"
	if r.Ref {
		verb = "references"
	} else if r.Fix {
		verb = "fixes"
	}
	msg := strings.TrimSpace(verb + " #" + strconv.Itoa(r.ID))
	if is.Title != nil && !r.Ref {
		msg = *is.Title + ", " + msg
	}
	log.Printf("loaded issue with msg %q", msg)
	c := exec.Command("git", "commit", "-e", "-m", msg)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c.Run()
}
