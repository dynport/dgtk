package main

import (
	"os"
	"strings"
	"time"

	"github.com/dynport/gocli"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	checkStatusCompleted   = "completed"
	checkStatusInProgress  = "in_progress"
	checkStatusQueued      = "queued"
	checkConclusionSuccess = "success"
	checkConclusionFailure = "failure"
)

type Checks struct {
	Branch string `cli:"opt --branch"`
	Wait   bool   `cli:"opt --wait"`
}

func (r *Checks) Run() error {
	l := logrus.New()
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

	if r.Wait {
		branch := r.Branch
		if branch == "" {
			branch, err = currentBranch()
			if err != nil {
				return errors.WithStack(err)
			}
		}

		completed := map[string]struct{}{}
		return waitFor(1*time.Second, 1*time.Hour, func() (bool, error) {
			res, err := loadChecks(cl, repo, branch)
			if err != nil {
				return false, errors.WithStack(err)
			}
			allCompleted := true
			for _, c := range res.CheckRuns {
				key := c.Name + c.Status
				if _, ok := completed[key]; !ok {
					l.Printf("%s %s %s %s", c.Name, c.Status, colorizeColclusion(c.Conclusion), c.DetailsURL)
					completed[key] = struct{}{}
				}
				if c.Status == checkStatusInProgress || c.Status == checkStatusQueued {
					allCompleted = false
				}
			}
			return allCompleted, nil

		})
	}

	added := 0
	runs := []*checkRun{}
	for _, branch := range branches {
		res, err := loadChecks(cl, repo, branch)
		if err != nil {
			return errors.WithStack(err)
		}
		runs = append(runs, res.CheckRuns...)
	}
	t := tablewriter.NewWriter(os.Stdout)
	for _, r := range runs {
		status := r.Status
		conclusion := colorizeColclusion(r.Conclusion)
		var completed string
		if r.CompletedAt != nil {
			completed = strings.Split(time.Since(*r.CompletedAt).String(), ".")[0]
		}
		added++
		t.Append([]string{r.Branch, r.Name, status, conclusion, completed, r.DetailsURL})
	}
	if added > 0 {
		t.Render()
	}
	return nil
}

func colorizeColclusion(conclusion string) string {
	if conclusion == checkConclusionSuccess {
		return gocli.Green(conclusion)
	} else {
		return gocli.Yellow(conclusion)
	}
	return ""
}

var ErrTimeout = errors.New("Timeout")

func waitFor(tick, max time.Duration, f func() (bool, error)) (err error) {
	t := time.NewTicker(tick)
	defer t.Stop()

	start := time.Now()
	var ok bool
	for _ = range t.C {
		ok, err = f()
		if ok {
			return nil
		} else if err != nil {
			return err
		}
		if time.Since(start) > max {
			break
		}
	}
	if err != nil {
		return err
	}
	return ErrTimeout
}
