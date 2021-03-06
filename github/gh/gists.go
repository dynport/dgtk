package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/dynport/dgtk/github"
	"github.com/dynport/gocli"
)

// https://developer.github.com/v3/gists/#delete-a-gist
var logger = log.New(os.Stderr, "", 0)

func githubToken() (token string, e error) {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}
	token, err := readGitConfig("github.token")
	if err != nil {
		return "", err
	} else if token == "" {
		return "", fmt.Errorf("unable to read github token")
	}
	return token, nil
}

func readGitConfig(name string) (string, error) {
	raw, e := exec.Command("git", "config", "--get", name).CombinedOutput()
	if e != nil {
		return "", fmt.Errorf("unable to get git config %q: %s", name, string(raw))
	}
	config := strings.TrimSpace(string(raw))
	if config == "" {
		return "", fmt.Errorf("no github config defined for %q", name)
	}
	return config, nil
}

func client() (*github.Client, error) {
	token, e := githubToken()
	if e != nil {
		return nil, e
	}
	return github.New(token)
}

func authenticatedRequest(method string, url string, r io.Reader) (*http.Response, error) {
	req, e := http.NewRequest(method, url, r)
	if e != nil {
		return nil, e
	}
	token, e := githubToken()
	if e != nil {
		return nil, e
	}
	dbg.Printf("using token %q", token)
	req.SetBasicAuth(string(token), "x-oauth-basic")
	dbg.Printf("sending request %q", req.URL)
	return http.DefaultClient.Do(req)
}

type DeleteGist struct {
	Id string `cli:"arg required"`
}

func (d *DeleteGist) Run() error {
	rsp, e := authenticatedRequest("DELETE", urlRoot+"/gists/"+d.Id, nil)
	if e != nil {
		return e
	}
	switch rsp.Status[0] {
	case '2':
		logger.Printf("deleted gist %q", d.Id)
		return nil
	default:
		b, _ := ioutil.ReadAll(rsp.Body)
		return fmt.Errorf("error deleteing gist status=%q error=%s response=%q", rsp.Status, e, string(b))
	}
}

type CreateGist struct {
	FileNames   []string `cli:"arg required"`
	Description string   `cli:"opt -d"`
	Public      bool     `cli:"opt -p --public"`
	Open        bool     `cli:"opt --open"`
}

func (g *CreateGist) Run() error {
	gist := &Gist{
		Description: g.Description,
		Public:      g.Public,
		Files:       Files{},
	}
	for _, name := range g.FileNames {
		b, e := ioutil.ReadFile(name)
		if e != nil {
			return e
		}
		gist.Files[path.Base(name)] = &File{
			Content: string(b),
		}
	}
	buf := &bytes.Buffer{}
	e := json.NewEncoder(buf).Encode(gist)
	if e != nil {
		return e
	}
	save := &bytes.Buffer{}
	tr := io.TeeReader(buf, save)
	url := urlRoot + "/gists"
	rsp, err := authenticatedRequest("POST", url, tr)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		b, _ := ioutil.ReadAll(rsp.Body)
		return fmt.Errorf("posting gist to url=%s: got status %s but expected 2x. body=%s", url, rsp.Status, string(b))
	}
	var rspGist *Gist
	if err := json.NewDecoder(rsp.Body).Decode(&rspGist); err != nil {
		return err
	}
	log.Printf("created gist %v", rspGist.HtmlUrl)
	if g.Open {
		return openUrl(rspGist.HtmlUrl)
	}
	return nil
}

type ListGists struct {
	Public bool `cli:"opt --public"`
}

const urlRoot = "https://api.github.com"

func (l *ListGists) Run() error {
	rsp, e := authenticatedRequest("GET", urlRoot+"/gists", nil)
	if e != nil {
		return e
	}
	defer rsp.Body.Close()
	gists := []*Gist{}
	e = json.NewDecoder(rsp.Body).Decode(&gists)
	if e != nil {
		return e
	}
	t := gocli.NewTable()
	for _, g := range gists {
		name := ""
		if len(g.Files) > 0 {
			for n := range g.Files {
				name = n
				break
			}
		}
		if g.Public || !l.Public {
			t.Add(g.CreatedAt.Local().Format(time.RFC3339), g.Id, name, g.Description)
		}
	}
	io.WriteString(os.Stdout, t.String()+"\n")
	return nil
}

type OpenGist struct {
	Id string `cli:"arg required"`
}

func validateSuccess(rsp *http.Response) error {
	switch rsp.Status[0] {
	case '2':
		return nil
	default:
		b, _ := ioutil.ReadAll(rsp.Body)
		return fmt.Errorf("error deleteing gist status=%q response=%q", rsp.Status, string(b))
	}
}

func (o *OpenGist) Run() error {
	rsp, e := authenticatedRequest("GET", urlRoot+"/gists/"+o.Id, nil)
	if e != nil {
		return e
	}
	defer rsp.Body.Close()
	e = validateSuccess(rsp)
	if e != nil {
		return e
	}
	gist := &Gist{}
	e = json.NewDecoder(rsp.Body).Decode(gist)
	if e != nil {
		return e
	}
	return openUrl(gist.HtmlUrl)
}

type BrowseGists struct {
}

func (b *BrowseGists) Run() error {
	githubUsername, e := readGitConfig("github.user")
	if e != nil {
		return e
	}

	return openUrl("https://gist.github.com/" + githubUsername)
}
