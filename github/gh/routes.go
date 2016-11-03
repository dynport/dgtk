package main

import "github.com/dynport/dgtk/cli"

func router() *cli.Router {
	router := cli.NewRouter()
	router.Register("browse", &Browse{}, "Browse github repository")
	router.Register("commits", &Commits{}, "List github commits")
	router.Register("gists/browse", &BrowseGists{}, "Browse Gists")
	router.Register("gists/create", &CreateGist{}, "Create a new")
	router.Register("gists/delete", &DeleteGist{}, "Create a new")
	router.Register("gists/list", &ListGists{}, "List Gists")
	router.Register("gists/open", &OpenGist{}, "Open a Gist")
	router.Register("issues/list", &issuesList{}, "List github issues")
	router.Register("issues/commit", &issuesCommit{}, "Commit an issue")
	router.Register("issues/browse", &issuesBrowse{}, "List github issues")
	router.Register("issues/create", &issuesCreate{}, "List github issues")
	router.Register("issues/open", &issueOpen{}, "Open github issues")
	router.Register("issues/label", &issueLabel{}, "Label issue")
	router.Register("issues/close", &issueClose{}, "Close github issues")
	router.Register("issues/assign", &issueAssign{}, "Assign gitbub issue")
	router.Register("notifications", &GithubNotifications{}, "Browse github notifications")
	router.Register("pulls", &GithubPulls{}, "List github pull requests")
	router.Register("teams/list", &teamsList{}, "Teams List")
	router.Register("teams/show", &teamsShow{}, "Teams Show")
	router.Register("teams/authorize", &teamsAuthorize{}, "Teams Authorize")
	router.Register("repos/list", &reposList{}, "Repos List")
	router.Register("repos/show", &reposShow{}, "Repos Show")
	router.Register("repos/create", &reposCreate{}, "Repos Create")
	router.Register("status", &Status{}, "Status")
	router.Register("clone", &cloneAction{}, "Clone Repo")
	return router
}
