package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dynport/gocli"
	"github.com/olekukonko/tablewriter"
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

	type checksResponse struct {
		CheckRuns []struct {
			Name        string     `json:"name"`
			ID          int        `json:"id"`
			Status      string     `json:"status"`
			Conclusion  string     `json:"conclusion"`
			URL         string     `json:"url"`
			DetailsURL  string     `json:"details_url"`
			StartedAt   *time.Time `json:"started_at"`
			CompletedAt *time.Time `json:"completed_at"`
		} `json:"check_runs"`
	}

	if r.Wait {
		return errors.Errorf("wait not implement yet")

	}

	t := tablewriter.NewWriter(os.Stdout)
	added := 0
	for _, branch := range branches {
		res, err := loadChecks(cl, repo, branch)
		if err != nil {
			return errors.WithStack(err)
		}
		var r *checksResponse
		err = json.Unmarshal(res, &r)
		if err != nil {
			return errors.WithStack(err)
		}
		for _, r := range r.CheckRuns {
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
			t.Append([]string{branch, r.Name, status, conclusion, completed, "https://github.com/phrase/x/runs/" + strconv.Itoa(r.ID)})
		}
	}
	if added > 0 {
		t.Render()
	}
	return nil
}
