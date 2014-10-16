package github

const (
	PullRequestStateClosed = "closed"
	PullRequestStateOpen   = "open"
	PullRequestStateAll    = "all"

	PullRequestSortCreated    = "created"
	PullRequestSortUpdated    = "updated"
	PullRequestSortPopularity = "popularity"

	PullRequestSortDesc = "desc"
	PullRequestSortAsc  = "asc"
)

type PullRequest struct {
	Url               string `json:"url"`                  // "https://api.github.com/repos/octocat/Hello-World/pulls/1",
	HtmlUrl           string `json:"html_url"`             // "https://github.com/octocat/Hello-World/pull/1",
	DiffUrl           string `json:"diff_url"`             // "https://github.com/octocat/Hello-World/pulls/1.diff",
	PatchUrl          string `json:"patch_url"`            // "https://github.com/octocat/Hello-World/pulls/1.patch"
	IssueUrl          string `json:"issue_url"`            // "https://api.github.com/repos/octocat/Hello-World/issues/1"
	CommitsUrl        string `json:"commits_url"`          // "https://api.github.com/repos/octocat/Hello-World/pulls/1/commits"
	ReviewCommentsUrl string `json:"review_comments_url"`  //  "https://api.github.com/repos/octocat/Hello-World/pulls/1/comments"
	ReviewCommentUrl  string `json:"review_comment_url"`   // "https://api.github.com/repos/octocat/Hello-World/pulls/comments/{number}"
	CommentsUrl       string `json:"comments_url"`         // "https://api.github.com/repos/octocat/Hello-World/issues/1/comments"
	StatusesUrl       string `json:"statuses_url"`         // "https://api.github.com/repos/octocat/Hello-World/statuses/6dcb09b5b57875f334f61aebed695e2e4193db5e"
	Number            int    `json:"number"`               // "1"
	State             string `json:"state"`                // "open"
	Title             string `json:"title"`                // "new-feature"
	Body              string `json:"body"`                 // "Please pull these awesome changes"
	CreatedAt         string `json:"created_at,omitempty"` // "2011-01-26T19:01:12Z",
	UpdatedAt         string `json:"updated_at,omitempty"` // "2011-01-26T19:14:43Z",
	ClosedAt          string `json:"closed_at,omitempty"`  // "2011-01-26T19:06:43Z",
	MergedAt          string `json:"merged_at,omitempty"`  // "2011-01-26T19:06:43Z",
}

type Head struct {
	Label string      `json:"label"` // "new-topic",
	Ref   string      `json:"ref"`   // "new-topic",
	Sha   string      `json:"sha"`   // "6dcb09b5b57875f334f61aebed695e2e4193db5e",
	User  *User       `json:"user"`
	Repo  *Repository `json:"repo"`
}

type Base struct {
	Label string      `json:"label"` // "new-topic",
	Ref   string      `json:"ref"`   // "new-topic",
	Sha   string      `json:"sha"`   // "6dcb09b5b57875f334f61aebed695e2e4193db5e",
	User  *User       `json:"user"`
	Repo  *Repository `json:"repo"`
}
