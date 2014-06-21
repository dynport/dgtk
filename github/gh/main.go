package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/dynport/dgtk/cli"
)

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
	for scanner := bufio.NewScanner(bytes.NewReader(out)); ; scanner.Scan() {
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
	for scanner := bufio.NewScanner(bytes.NewReader(out)); ; scanner.Scan() {
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
	c := exec.Command("open", theUrl)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = nil
	return c.Run()
}

var router = cli.NewRouter()

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

func main() {
	log.SetFlags(0)
	router.Register("browse", &Browse{}, "Browse github repository")
	router.Register("commits", &Commits{}, "List github commits")
	router.Register("gists/browse", &BrowseGists{}, "Browse Gists")
	router.Register("gists/create", &CreateGist{}, "Create a new")
	router.Register("gists/delete", &DeleteGist{}, "Create a new")
	router.Register("gists/list", &ListGists{}, "List Gists")
	router.Register("gists/open", &OpenGist{}, "Open a Gist")
	router.Register("issues/list", &issuesList{}, "List github issues")
	router.Register("issues/browse", &issuesBrowse{}, "List github issues")
	router.Register("issues/create", &issuesCreate{}, "List github issues")
	router.Register("issues/open", &issueOpen{}, "Open github issues")
	router.Register("issues/tag", &issueTag{}, "Tag issue")
	router.Register("issues/close", &issueClose{}, "Close github issues")
	router.Register("issues/assign", &issueAssign{}, "Assign gitbub issue")
	router.Register("notifications", &GithubNotifications{}, "Browse github notifications")
	router.Register("pulls", &GithubPulls{}, "List github pull requests")
	e := router.RunWithArgs()
	switch e {
	case nil, cli.ErrorHelpRequested, cli.ErrorNoRoute:
		// ignore
	default:
		log.Fatal(e.Error())
	}
}
