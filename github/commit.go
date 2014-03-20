package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"
)

type ListCommits struct {
	Org    string
	Repo   string
	Sha    string
	Path   string
	Author string
	Since  time.Time
	Until  time.Time
}

func (a *ListCommits) Execute(client *Client) ([]*Commit, error) {
	if a.Org == "" {
		return nil, fmt.Errorf("Org must be set")
	}
	if a.Repo == "" {
		return nil, fmt.Errorf("Repo must be set")
	}
	v := url.Values{}
	if a.Sha != "" {
		v.Set("sha", a.Sha)
	}
	if a.Path != "" {
		v.Set("path", a.Path)
	}

	if a.Author != "" {
		v.Set("author", a.Author)
	}

	if !a.Since.IsZero() {
		v.Set("since", a.Since.UTC().Format("2006-01-02T15:04:05.00Z"))
	}
	if !a.Until.IsZero() {
		v.Set("until", a.Until.UTC().Format("2006-01-02T15:04:05.00Z"))
	}
	theUrl := "https://api.github.com/repos/" + a.Org + "/" + a.Repo + "/commits"
	if len(v) > 0 {
		theUrl += "?" + v.Encode()
	}
	rsp, e := client.Get(theUrl)

	if e != nil {
		return nil, e
	}
	if rsp.Status[0] != '2' {
		return nil, fmt.Errorf("expected status 2xx, got %s", rsp.Status)
	}
	defer rsp.Body.Close()

	commits := []*Commit{}
	buf := &bytes.Buffer{}
	tee := io.TeeReader(rsp.Body, buf)
	e = json.NewDecoder(tee).Decode(&commits)
	return commits, e
}

type Tree struct {
	Url string `json:"url"` // "https://api.github.com/repos/octocat/Hello-World/tree/6dcb09b5b57875f334f61aebed695e2e4193db5e",
	Sha string `json:"sha"` // "6dcb09b5b57875f334f61aebed695e2e4193db5e"
}

type CommitUserDetails struct {
	Name  string    `json:"name"`  // "Monalisa Octocat",
	Email string    `json:"email"` // "support@github.com",
	Date  time.Time `json:"date"`  // "2011-04-14T16:00:49Z"
}

type CommitDetails struct {
	Url       string             `json:"url"`
	Author    *CommitUserDetails `json:"author"`
	Committer *CommitUserDetails `json:"committer"`
	Message   string             `json:"message"` // "Fix all the bugs",
	Tree      *Tree              `json:"tree"`
}

type Branch struct {
	Name   string  `json:"name"`
	Commit *Commit `json:"commit"`
	Links  *Links  `json:"_links"`
}

type Links struct {
	Html string `json:"html"`
	Self string `json:"self"`
}

type Commit struct {
	Url       string         `json:"url"` // "https://api.github.com/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e",
	Sha       string         `json:"sha"` // "6dcb09b5b57875f334f61aebed695e2e4193db5e",
	Commit    *CommitDetails `json:"commit"`
	Author    *User          `json:"author"`
	Committer *User          `json:"committer"`
	Parents   []*Tree        `json:"parents"`
}

type User struct {
	Login             string `json:"login"`               // "octocat",
	Id                int64  `json:"id"`                  // 1,
	AvatarUrl         string `json:"avatar_url"`          // "https://github.com/images/error/octocat_happy.gif",
	GravatarId        string `json:"gravatar_id"`         // "somehexcode",
	Url               string `json:"url"`                 // "https://api.github.com/users/octocat",
	HtmlUrl           string `json:"html_url"`            // "https://github.com/octocat",
	FollowersUrl      string `json:"followers_url"`       // "https://api.github.com/users/octocat/followers",
	FollowingUrl      string `json:"following_url"`       // "https://api.github.com/users/octocat/following{/other_user}",
	GistsUrl          string `json:"gists_url"`           // "https://api.github.com/users/octocat/gists{/gist_id}",
	StarredUrl        string `json:"starred_url"`         // "https://api.github.com/users/octocat/starred{/owner}{/repo}",
	SubscriptionsUrl  string `json:"subscriptions_url"`   // "https://api.github.com/users/octocat/subscriptions",
	OrganizationsUrl  string `json:"organizations_url"`   // "https://api.github.com/users/octocat/orgs",
	ReposUrl          string `json:"repos_url"`           // "https://api.github.com/users/octocat/repos",
	EventsUrl         string `json:"events_url"`          // "https://api.github.com/users/octocat/events{/privacy}",
	ReceivedEventsUrl string `json:"received_events_url"` // "https://api.github.com/users/octocat/received_events",
	Type              string `json:"type"`                // "User",
	SiteAdmin         bool   `json:"site_admin"`          // false
}
