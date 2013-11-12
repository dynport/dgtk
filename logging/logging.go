package logging

import (
	"compress/gzip"
	"io"
	"os/exec"
	"time"
)

type RemoteLog struct {
	Host          string
	Pattern       string
	Tail          bool
	Time          time.Time
	Compress      bool
	CustomLogRoot string
}

const (
	DEFAULT_LOG_ROOT = "/var/log/hourly"
	HOURLY_PATTERN   = "2006/01/02/2006-01-02T15.log"
)

func NewRemoteLogFromTime(host string, t time.Time, pattern string) *RemoteLog {
	return &RemoteLog{
		Time:    t,
		Host:    host,
		Pattern: pattern,
	}
}

func (rl *RemoteLog) LogRoot() string {
	if rl.CustomLogRoot != "" {
		return rl.CustomLogRoot
	}
	return DEFAULT_LOG_ROOT
}

func (rl *RemoteLog) Current() string {
	return rl.LogRoot() + "/current"
}

func (rl *RemoteLog) Path() string {
	if !rl.Time.IsZero() {
		return rl.Time.UTC().Format(rl.LogRoot() + "/" + HOURLY_PATTERN)
	}
	return rl.Current()
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
		return "tail -n 0 -F " + rl.Current()
	}
	return "{ test -e " + rl.Path() + " && cat " + rl.Path() + "; test -e " + rl.GzipPath() + " && cat " + rl.GzipPath() + " | gunzip; }"
}

func (rl *RemoteLog) Reader() (reader io.ReadCloser, e error) {
	c := rl.Command()
	var cmd *exec.Cmd
	if rl.Host != "" {
		cmd = exec.Command("ssh", "-t", "-l", "root", rl.Host, c)
	} else {
		cmd = exec.Command("bash", "-c", c)
	}
	reader, e = cmd.StdoutPipe()
	if e != nil {
		return nil, e
	}
	e = cmd.Start()
	if e != nil {
		return nil, e
	}
	if rl.Compress {
		reader, e = gzip.NewReader(reader)
		if e != nil {
			return nil, e
		}
	}
	return reader, nil
}
