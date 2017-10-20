package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/dynport/dgtk/github"
	"github.com/dynport/gocli"
)

type Status struct {
	WithURLs bool   `cli:"opt --with-urls"`
	Branch   string `cli:"opt --branch"`
	Wait     bool   `cli:"opt --wait"`
}

func (r *Status) Run() error {

	if r.Wait {
	}
	var branches []string
	if r.Branch != "" {
		branches = []string{r.Branch}
	} else {
		branches = []string{"master"}
		if cb, err := currentBranch(); err == nil && cb != "master" {
			branches = append([]string{cb}, "master")
		}
	}
	repo, err := githubRepo()
	if err != nil {
		return err
	}
	cl, err := client()
	if err != nil {
		return err
	}

	l := log.New(os.Stderr, "", 0)

	if r.Wait {
		if r.Branch == "" {
			return fmt.Errorf("wait requires one specific branch")
		}
		var printedURL bool
		for {
			s, err := loadStatus(cl, repo, r.Branch)
			if err != nil {
				l.Printf("error fetching status: %s", err)
			} else {
				if s.State != statePending {
					fmt.Println(s.State)
					if s.State == stateSuccess {
						return nil
					}
					return fmt.Errorf("not successful (%s)", s.State)
				}
				if !printedURL && len(s.Statuses) > 0 {
					l.Printf("url=%s", s.Statuses[0].TargetURL)
					printedURL = true
				}
			}
			time.Sleep(10 * time.Second)
		}
		return nil
	}

	t := gocli.NewTable()
	for _, b := range branches {
		var ago, url, status, sha string
		if s, err := loadStatus(cl, repo, b); err != nil {
			if isNotFound(err) {
				status = "not_found"
			} else {
				return err
			}
		} else {
			status = s.State
			sha = s.SHA
			if len(s.Statuses) > 0 {
				ago = strings.Split(time.Since(s.Statuses[0].CreatedAt).String(), ".")[0]
				url = s.Statuses[0].TargetURL
			}
		}
		args := []interface{}{b, colorizeStatus(status), truncate(sha, 8, false), ago}
		if r.WithURLs {
			args = append(args, url)
		}
		t.Add(args...)
	}
	fmt.Println(t)
	return nil
}

func isNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "404 Not Found")
}

const (
	stateSuccess  = "success"
	statePending  = "pending"
	stateNotFound = "not_found"
)

func colorizeStatus(in string) string {
	color := gocli.Green
	switch in {
	case stateSuccess:
		color = gocli.Green
	case statePending, stateNotFound:
		color = gocli.Yellow
	default:
		color = gocli.Red
	}
	return color(in)
}

func loadStatus(cl *github.Client, repo, ref string) (res *statusResponse, err error) {
	u := "https://api.github.com/repos/" + repo + "/commits/" + ref + "/status"
	rsp, err := cl.Get(u)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		b, _ := ioutil.ReadAll(rsp.Body)
		return nil, fmt.Errorf("got status %s but expected 2x. body=%s", rsp.Status, string(b))
	}
	err = json.NewDecoder(rsp.Body).Decode(&res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func currentBranch() (string, error) {
	b, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

type statusResponse struct {
	State    string `json:"state"`
	Statuses []*struct {
		URL       string    `json:"url,omitempty"`
		TargetURL string    `json:"target_url,omitempty"`
		CreatedAt time.Time `json:"created_at,omitempty"`
		UpdatedAt time.Time `json:"updated_at,omitempty"`
	} `json:"statuses"`
	SHA string `json:"sha"`
}

// to be used to colorize
func dataOn(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}
