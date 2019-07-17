package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/dynport/dgtk/github"
	"github.com/dynport/gocli"
	"github.com/pkg/errors"
)

type Status struct {
	Open   bool   `cli:"opt --open"`
	Branch string `cli:"opt --branch"`
	Wait   bool   `cli:"opt --wait"`
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
				if s.State != statusPending {
					fmt.Println(s.State)
					if s.State == statusSuccess {
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
			if len(sm) == 0 {
				st.Status = s.State
			} else {
				if sm[statusFailure] > 0 {
					st.Status = statusFailure
					failures++
				} else if sm[statusPending] > 0 {
					st.Status = statusPending
				} else {
					st.Status = statusSuccess
				}
			}
			t.Add(string(b), colorizeStatus(st.Status))
			if len(s.Statuses) > 0 {
				for _, ss := range s.Statuses {
					if ss.TargetURL != "" {
						st.URLs = append(st.URLs, ss.TargetURL)
					}
					t.Add("", colorizeStatus(ss.State), truncate(s.SHA, 8, false), ss.Context, agoFunc(ss.CreatedAt), ss.TargetURL)
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
	statusSuccess  = "success"
	statusPending  = "pending"
	statusNotFound = "not_found"
	statusFailure  = "failure"
)

func colorizeStatus(in string) string {
	color := gocli.Green
	switch in {
	case statusSuccess:
		color = gocli.Green
	case statusPending, statusNotFound:
		color = gocli.Yellow
	default:
		color = gocli.Red
	}
	return color(in)
}

type checkRun struct {
	Name        string     `json:"name"`
	Branch      string     `json:"branch"`
	ID          int        `json:"id"`
	Status      string     `json:"status"`
	Conclusion  string     `json:"conclusion"`
	URL         string     `json:"url"`
	DetailsURL  string     `json:"details_url"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type checksResponse struct {
	CheckRuns []*checkRun `json:"check_runs"`
}

func loadChecks(cl *github.Client, repo, ref string) (res *checksResponse, err error) {
	req, err := http.NewRequest("GET", "https://api.github.com/repos/"+repo+"/commits/"+ref+"/check-runs", nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.Header.Set("Accept", "application/vnd.github.antiope-preview+json")

	rsp, err := cl.Do(req)
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
	for _, r := range res.CheckRuns {
		r.Branch = ref
	}
	return res, nil
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
	TotalCount int    `json:"total_count"`
	State      string `json:"state"`
	Statuses   []*struct {
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
