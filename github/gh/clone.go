package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type cloneAction struct {
	Repo string `cli:"arg required"`
	Dst  string `cli:"opt -d --dst default='$HOME/src/github.com'"`
}

func (r *cloneAction) Run() error {
	l := log.New(os.Stderr, "", 0)
	withoutGit := strings.TrimSuffix(r.Repo, ".git")
	withGit := withoutGit + ".git"
	dir := os.ExpandEnv(r.Dst + "/" + withoutGit)
	if stat, err := os.Stat(dir); err == nil {
		if stat.IsDir() {
			l.Printf("repo already cloned to %s", dir)
		} else {
			l.Printf("repo not cloned but a file exists at %s", dir)
		}
		return nil
	}
	err := os.MkdirAll(filepath.Dir(dir), 0755)
	if err != nil {
		return err
	}
	repo := "git@github.com:" + withGit
	l.Printf("cloning %s to %s", repo, dir)
	c := exec.Command("git", "clone", repo, dir)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err = c.Run()
	if err != nil {
		return err
	}
	l.Printf("cloned into %s", dir)
	return nil
}

func cloneRepo(src, dst string) error {
	return nil
}
