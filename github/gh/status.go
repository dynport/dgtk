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
	"github.com/pkg/errors"
)

type Status struct {
	WithURLs bool   `cli:"opt --with-urls"`
	Open     bool   `cli:"opt --open"`
	Branch   string `cli:"opt --branch"`
	Wait     bool   `cli:"opt --wait"`
}

func (r *Status) Run() error {
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
		branch := r.Branch
		if branch == "" {
			branch, err = currentBranch()
			if err != nil {
				return err
			}
		}
		var printedURL bool
		for {
			s, err := loadStatus(cl, repo, branch)
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

	type status struct {
		Time   time.Time
		Branch string
		URL    string
		URLs   []string
		Status string
		SHA    string
	}

	t := gocli.NewTable()
	all := []*status{}
	failures := 0
	agoFunc := func(t time.Time) string { return strings.Split(time.Since(t).String(), ".")[0] }
	for _, b := range branches {
		st := &status{Branch: b}
		all = append(all, st)
		if s, err := loadStatus(cl, repo, b); err != nil {
			if isNotFound(err) {
				st.Status = "not_found"
			} else {
				return err
			}
		} else {
			st.Status = s.State
			st.SHA = s.SHA
			sm := map[string]int{}
			for _, s := range s.Statuses {
				sm[s.State]++
			}
			if sm["failure"] > 0 {
				st.Status = "failure"
				failures++
			} else if sm["pending"] > 0 {
				st.Status = "pending"
			} else {
				st.Status = "success"
			}
			t.Add(string(b), colorizeStatus(st.Status))
			if len(s.Statuses) > 0 {
				for _, ss := range s.Statuses {
					if ss.TargetURL != "" {
						st.URLs = append(st.URLs, ss.TargetURL)
					}
					args := []interface{}{"", colorizeStatus(ss.State), truncate(s.SHA, 8, false), ss.Context, agoFunc(ss.CreatedAt)}
					if r.WithURLs {
						args = append(args, ss.TargetURL)
					}
					t.Add(args...)
				}
			}
		}
	}

	if r.Open {
		if len(all) == 0 {
			return fmt.Errorf("no status found")
		}
		s := all[0]
		for _, s := range all {
			l.Printf("url: %s", s.URL)
		}
		url := ""
		if s.URL != "" {
			url = s.URL
		} else if len(s.URLs) == 1 {
			url = s.URLs[0]
		} else {
			return fmt.Errorf("status has no url (yet?). url=%q urls=%#v", s.URL, s.URLs)
		}
		return openUrl(url)
	}
	fmt.Println(t)
	if failures > 0 {
		return errors.Errorf("%d failures", failures)
	}
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
		State     string    `json:"state,omitempty"`
		URL       string    `json:"url,omitempty"`
		Context   string    `json:"context"`
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
