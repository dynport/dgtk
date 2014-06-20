package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/dynport/gocli"
)

type issuesList struct {
	All bool `cli:"opt --all"`
}

func (r *issuesList) Run() error {
	var e error
	repo := ""
	if !r.All {
		repo, e = githubRepo()
		if e != nil {
			logger.Printf("ERROR=%q", e)
		}
	}
	issues, e := loadIssues(repo)
	if e != nil {
		return e
	}
	if len(issues) == 0 {
		fmt.Println("no issues found")
		return nil
	}
	t := gocli.NewTable()
	for _, i := range issues {
		orga, issueRepo, e := i.Repo()
		if e != nil {
			return e
		}
		labels := []string{}
		for _, l := range i.Labels {
			labels = append(labels, l.Name)
		}
		parts := []interface{}{i.Number}
		if repo == "" {
			parts = append(parts, orga+"/"+issueRepo)
		}
		assignee := ""
		if i.Assignee != nil {
			assignee = i.Assignee.Login
		}
		parts = append(parts, truncate(i.CreatedAt, 16, false), i.State, assignee, truncate(i.Title, 48, true), strings.Join(labels, ","))
		t.Add(parts...)
	}
	fmt.Println(t)
	return nil
}

func truncate(s string, l int, dots bool) string {
	if len(s) > l {
		if l > 6 && dots {
			return s[0:l-3] + "..."
		}
		return s[0:l]
	}
	return s
}

func loadIssues(repo string) ([]*Issue, error) {
	u := urlRoot + "/issues"
	if repo != "" {
		u = urlRoot + "/repos/" + repo + "/issues"
	}
	dbg.Printf("listing issues for %q", u)
	rsp, e := authenticatedRequest("GET", u, nil)
	if e != nil {
		return nil, e
	}
	defer rsp.Body.Close()
	b, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return nil, e
	}
	if rsp.Status[0] != '2' {
		return nil, fmt.Errorf("expected status 2xx, got %s: %s", rsp.Status, string(b))
	}
	issues := []*Issue{}
	e = json.Unmarshal(b, &issues)
	if e != nil {
		logger.Printf("%s", string(b))
	}
	return issues, e
}

// https://developer.github.com/v3/issues/#create-an-issue

type CreateIssue struct {
	Title     string   `json:"title,omitempty"`
	Body      string   `json:"body,omitempty"`
	Assignee  string   `json:"assignee,omitempty"`
	Milestone int      `json:"milestone,omitempty"`
	State     string   `json:"state,omitempty"`
	Labels    []string `json:"labels,omitempty"`
}

type Issue struct {
	Url         string       `json:"url,omitempty"`      // "https://api.github.com/repos/octocat/Hello-World/issues/1347",
	HtmlUrl     string       `json:"html_url,omitempty"` // "https://github.com/octocat/Hello-World/issues/1347",
	Number      int          `json:"number,omitempty"`   // 1347,
	State       string       `json:"state,omitempty"`    // "open",
	Title       string       `json:"title,omitempty"`    // "Found a bug",
	Body        string       `json:"body,omitempty"`     // "I'm having a problem with this.",
	User        *User        `json:"user,omitempty"`
	Assignee    *User        `json:"assignee,omitempty"`
	Labels      []*Label     `json:"labels,omitempty"`
	Milestone   *Milestone   `json:"milestone,omitempty"`
	Commens     int          `json:"comments,omitempty"`
	PullRequest *PullRequest `json:"pull_request,omitempty"`
	ClosedAt    string       `json:"closed_at,omitempty"`  // null,
	CreatedAt   string       `json:"created_at,omitempty"` // "2011-04-22T13:33:48Z",
	UpdatedAt   string       `json:"updated_at,omitempty"` // "2011-04-22T13:33:48Z"
}

var reposPreifx = "https://api.github.com/repos/"

func (i *Issue) Repo() (string, string, error) {
	if strings.HasPrefix(i.Url, reposPreifx) {
		parts := strings.Split(strings.TrimPrefix(i.Url, reposPreifx), "/")
		if len(parts) > 2 {
			return parts[0], parts[1], nil
		}
	}
	return "", "", fmt.Errorf("unable to extract repo from %q", i.Url)
}

type PullRequest struct {
	Url      string `json:"url"`       // "https://api.github.com/repos/octocat/Hello-World/pulls/1347",
	HtmlUrl  string `json:"html_url"`  // "https://github.com/octocat/Hello-World/pull/1347",
	DiffUrl  string `json:"diff_url"`  // "https://github.com/octocat/Hello-World/pull/1347.diff",
	PatchUrl string `json:"patch_url"` // "https://github.com/octocat/Hello-World/pull/1347.patch"
}

type Label struct {
	Url   string `json:"url"`   // "https://api.github.com/repos/octocat/Hello-World/labels/bug",
	Name  string `json:"name"`  // "bug",
	Color string `json:"color"` // "f29513"
}

type Milestone struct {
	Url          string `json:"url"`         // "https://api.github.com/repos/octocat/Hello-World/milestones/1",
	Number       int    `json:"number"`      // 1,
	State        string `json:"state"`       // "open",
	Title        string `json:"title"`       // "v1.0",
	Description  string `json:"description"` // "",
	Creator      *User  `json:"creator"`
	OpenIssues   int    `json:"open_issues"`          // 4,
	ClosedIssues int    `json:"closed_issues"`        // 8,
	CreatedAt    string `json:"created_at,omitempty"` // "2011-04-10T20:09:31Z",
	UpdatedAt    string `json:"updated_at,omitempty"` // "2014-03-03T18:58:10Z",
	DueOn        string `json:"due_on,omitempty"`     // null
}
