package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// http://developer.github.com/v3/repos/contents/#get-contents
type GetContents struct {
	Owner string
	Repo  string
	Path  string
	Ref   string
}

func (action *GetContents) Execute(client *Client) ([]*Content, error) {
	theUrl := "https://api.github.com/repos/" + action.Owner + "/" + action.Repo + "/contents"
	if action.Path != "" {
		theUrl += "/" + action.Path
	}
	if action.Ref != "" {
		theUrl += "?ref=" + action.Ref
	}
	rsp, e := client.Get(theUrl)
	if e != nil {
		return nil, e
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		return nil, fmt.Errorf("expected status 2xx, got %v", rsp.Status)
	}
	b, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return nil, e
	}

	// this is a brilliant api
	var c *Content
	e = json.Unmarshal(b, &c)
	if e == nil {
		return []*Content{c}, nil
	}
	var cs []*Content
	return cs, json.Unmarshal(b, &cs)
}

type Content struct {
	Type    string `json:"type"`     // "file",
	Size    int64  `json:"size"`     // 625,
	Name    string `json:"name"`     // "octokit.rb",
	Path    string `json:"path"`     // "lib/octokit.rb",
	Sha     string `json:"sha"`      // "fff6fe3a23bf1c8ea0692b4a883af99bee26fd3b",
	Url     string `json:"url"`      // "https://api.github.com/repos/pengwynn/octokit/contents/lib/octokit.rb",
	GitUrl  string `json:"git_url"`  // "https://api.github.com/repos/pengwynn/octokit/git/blobs/fff6fe3a23bf1c8ea0692b4a883af99bee26fd3b",
	HtmlUrl string `json:"html_url"` // "https://github.com/pengwynn/octokit/blob/master/lib/octokit.rb",
	Links   *Links `json:"_links"`
}
