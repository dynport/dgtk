package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/dynport/dgtk/cli"
)

func main() {
	log.SetFlags(0)
	r := router()
	e := r.RunWithArgs()
	switch e {
	case nil, cli.ErrorHelpRequested, cli.ErrorNoRoute:
		// ignore
	default:
		log.Fatal(e.Error())
	}
}

type Commits struct {
}

func (c *Commits) Run() error {
	theUrl, e := githubUrl()
	if e != nil {
		return e
	}
	return openUrl(theUrl + "/commits/master")
}

func githubRepo() (string, error) {
	out, e := exec.Command("git", "remote", "-v").CombinedOutput()
	if e != nil {
		if strings.Contains(string(out), "Not a git repository") {
			return "", nil
		}
		return "", fmt.Errorf("%s: %s:", e, string(out))
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) > 1 && strings.HasPrefix(fields[1], "git@github.com:") {
			repo := fields[1]
			parts := strings.Split(repo, ":")
			return strings.TrimSuffix(parts[1], ".git"), nil
		}
	}
	return "", e

}

func githubUrl() (string, error) {
	out, e := exec.Command("git", "remote", "-v").CombinedOutput()
	if e != nil {
		return "", e
	}
	for scanner := bufio.NewScanner(bytes.NewReader(out)); scanner.Scan(); {
		fields := strings.Fields(scanner.Text())
		if len(fields) > 1 && strings.HasPrefix(fields[1], "git@github.com:") {
			repo := fields[1]
			parts := strings.Split(repo, ":")
			if len(parts) > 1 {
				return "https://github.com/" + strings.TrimSuffix(parts[1], ".git"), nil
			}
		}
	}
	return "", fmt.Errorf("error getting github url from %s (I only know about 'git@github.com/' remotes for now", string(out))
}

type Browse struct {
}

func (o *Browse) Run() error {
	theUrl, e := githubUrl()
	if e != nil {
		return e
	}
	return openUrl(theUrl)
}

func openGithubUrl(suffix string) error {
	u, e := githubUrl()
	if e != nil {
		return e
	}
	u += "/" + strings.TrimPrefix(suffix, "/")
	return openUrl(u)
}

func openUrl(theUrl string) error {
	logger.Printf("opening %q", theUrl)
	for _, n := range []string{"xdg-open", "open"} {
		if p, err := exec.LookPath(n); err == nil {
			c := exec.Command(p, theUrl)
			err := c.Start()
			if err == nil {
				return nil
			}
			logger.Printf("err=%q", err)
		}
	}
	return fmt.Errorf("could not find command to open url")
}

type GithubNotifications struct {
}

func (g *GithubNotifications) Run() error {
	return openUrl("https://github.com/notifications")
}

type GithubPulls struct {
}

func (g *GithubPulls) Run() error {
	u, e := githubUrl()
	if e != nil {
		return e
	}
	return openUrl(u + "/pulls")
}
