package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dynport/gocli"
	"github.com/olekukonko/tablewriter"
	"github.com/phrase/x/wait"
	"github.com/pkg/errors"
)

const (
	checkStatusCompleted   = "completed"
	checkStatusInProgress  = "in_progress"
	checkConclusionSuccess = "success"
	checkConclusionFailure = "failure"
)

type Checks struct {
	Branch string `cli:"opt --branch"`
	Wait   bool   `cli:"opt --wait"`
}

func (r *Checks) Run() error {
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

	printCheckRuns := func(runs []*checkRun) {
		return
	}

	if r.Wait {
		branch := r.Branch
		if branch == "" {
			branch, err = currentBranch()
			if err != nil {
				return errors.WithStack(err)
			}
		}

		return wait.For(1*time.Second, 1*time.Hour, func() (bool, error) {
			res, err := loadChecks(cl, repo, branch)
			if err != nil {
				return false, errors.WithStack(err)
			}
			allCompleted := func() bool {
				for _, c := range res.CheckRuns {
					if c.Status == checkStatusCompleted {
						return true
					}
				}
				return false
			}()
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
		conclusion := r.Conclusion
		if conclusion == checkConclusionSuccess {
			conclusion = gocli.Green(conclusion)
		} else {
			conclusion = gocli.Yellow(conclusion)
		}
		var completed string
		if r.CompletedAt != nil {
			completed = strings.Split(time.Since(*r.CompletedAt).String(), ".")[0]
		}
		added++
		t.Append([]string{r.Branch, r.Name, status, conclusion, completed, "https://github.com/" + repo + "/runs/" + strconv.Itoa(r.ID)})
	}
	if added > 0 {
		t.Render()
	}
	return nil
}
