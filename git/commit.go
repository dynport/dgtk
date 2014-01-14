package git

import (
	"time"
)

type Commit struct {
	Checksum   string
	AuthorDate time.Time
	Message    string
}

type CommitOptions struct {
	Limit   int
	Pattern string
}
