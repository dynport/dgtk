package main

import "time"

type Gist struct {
	CommentsUrl string    `json:"comments_url,omitempty"`
	Description string    `json:"description,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	Public      bool      `json:"public,omitempty"`
	Url         string    `json:"url,omitempty"`
	ForksUrl    string    `json:"forks_url,omitempty"`
	CommitsUrl  string    `json:"commits_url,omitempty"`
	Id          string    `json:"id,omitempty"`
	GitPullUrl  string    `json:"git_pull_url,omitempty"`
	GitPushUrl  string    `json:"git_push_url,omitempty"`
	HtmlUrl     string    `json:"html_url,omitempty"`
	Comments    int64     `json:"comments,omitempty"`
	User        *User     `json:"user,omitempty"`
	Files       Files     `json:"files,omitempty"`
}

type User struct {
	SiteAdmin         bool   `json:"site_admin,omitempty"`
	FollowingUrl      string `json:"following_url,omitempty"`
	FollowersUrl      string `json:"followers_url,omitempty"`
	HtmlUrl           string `json:"html_url,omitempty"`
	Url               string `json:"url,omitempty"`
	GravatarId        string `json:"gravatar_id,omitempty"`
	AvatarUrl         string `json:"avatar_url,omitempty"`
	Id                int64  `json:"id,omitempty"`
	Login             string `json:"login,omitempty"`
	GistsUrl          string `json:"gists_url,omitempty"`
	StarredUrl        string `json:"starred_url,omitempty"`
	SubscriptionsUrl  string `json:"subscriptions_url,omitempty"`
	OrganizationsUrl  string `json:"organizations_url,omitempty"`
	ReposUrl          string `json:"repos_url,omitempty"`
	EventsUrl         string `json:"events_url,omitempty"`
	ReceivedEventsUrl string `json:"received_events_url,omitempty"`
	Type              string `json:"type,omitempty"`
}

type Files map[string]*File

type File struct {
	Size     int64  `json:"size,omitempty"`
	RawUrl   string `json:"raw_url,omitempty"`
	Language string `json:"language,omitempty"`
	Type     string `json:"type,omitempty"`
	Filename string `json:"filename,omitempty"`
	Content  string `json:"content,omitempty"`
}
