package github

import (
	"fmt"
	"strings"
)

const (
	IssueStateClosed = "closed"
	IssueStateOpen   = "open"
	IssueStateAll    = "all"

	IssueSortCreated  = "created"
	IssueSortUpdated  = "updated"
	IssueSortComments = "comments"

	IssueAssigneeNonde = "none"
	IssueAssigneeAny   = "*"

	IssueSortDesc = "desc"
	IssueSortAsc  = "asc"
)

var reposPreifx = "https://api.github.com/repos/"

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

func (i *Issue) Repo() (string, string, error) {
	if strings.HasPrefix(i.Url, reposPreifx) {
		parts := strings.Split(strings.TrimPrefix(i.Url, reposPreifx), "/")
		if len(parts) > 2 {
			return parts[0], parts[1], nil
		}
	}
	return "", "", fmt.Errorf("unable to extract repo from %q", i.Url)
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
