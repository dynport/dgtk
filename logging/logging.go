package logging

import (
	"bufio"
	"compress/gzip"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"
)

type RemoteLog struct {
	Host     string
	Pattern  string
	Tail     bool
	Time     time.Time
	Compress bool
}

const (
	LOG_ROOT       = "/var/log/hourly"
	CURRENT_ROOT   = LOG_ROOT + "/current"
	HOURLY_PATTERN = "2006/01/02/2006-01-02T15.log"
)

func NewRemoteLogFromTime(host string, t time.Time, pattern string) *RemoteLog {
	return &RemoteLog{
		Time:    t,
		Host:    host,
		Pattern: pattern,
	}
}

func (rl *RemoteLog) Path() string {
	if !rl.Time.IsZero() {
		return rl.Time.Format(LOG_ROOT + "/" + HOURLY_PATTERN)
	}
	return CURRENT_ROOT
}

func (rl *RemoteLog) GzipPath() string {
	return rl.Path() + ".gz"
}

func (rl *RemoteLog) Command() string {
	cmd := rl.CatCmd()
	if rl.Pattern != "" {
		cmd += " | " + rl.GrepCmd()
	}
	if rl.Compress {
		cmd += " | gzip"
	}
	return cmd
}

func (rl *RemoteLog) GrepCmd() string {
	return "grep " + rl.Pattern
}

func (rl *RemoteLog) CatCmd() string {
	if rl.Tail {
		return "tail -n 0 -F " + CURRENT_ROOT
	}
	return "{ test -e " + rl.Path() + " && cat " + rl.Path() + "; test -e " + rl.GzipPath() + " && cat " + rl.GzipPath() + " | gunzip; }"
}

func (rl *RemoteLog) Reader() (reader io.ReadCloser, e error) {
	c := rl.Command()
	log.Println("CMD: " + c)
	cmd := exec.Command("ssh", "-t", "-l", "root", rl.Host, c)
	reader, e = cmd.StdoutPipe()
	if e != nil {
		return nil, e
	}
	e = cmd.Start()
	if e != nil {
		return nil, e
	}
	if rl.Compress {
		log.Println("making logger compressed")
		reader, e = gzip.NewReader(reader)
		if e != nil {
			return nil, e
		}
		log.Println("made logger compressed")
	}
	return reader, nil
}
