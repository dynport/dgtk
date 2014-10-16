package github

import (
	"encoding/json"
	"testing"
)

func TestPullRequest(t *testing.T) {
	pr := &PullRequest{}
	err := json.Unmarshal([]byte(s), &pr)
	if err != nil {
		t.Fatalf("expected no error unmarshaling, got %q", err)
	}
	created := pr.CreatedAt.Format("2006-01-02")
	if created != "2011-01-26" {
		t.Errorf("expected CreatedAt to equal %q, was %q", "2011-01-26", created)
	}
}

const s = `{"created_at":"2011-01-26T19:01:12Z"}`
