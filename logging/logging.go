package logging

import (
	"compress/gzip"
	"io"
	"os/exec"
	"time"
)

type RemoteLog struct {
	Host          string
	User          string
	Pattern       string
	Tail          bool
	Time          time.Time
	Compress      bool
	CustomLogRoot string
	FromBegin     bool // to be used with tail
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
		n := "0"
		if rl.FromBegin {
			n = "+0"
		}
		return "tail -n " + n + " -F " + rl.Current()
	}
	return "{ test -e " + rl.Path() + " && cat " + rl.Path() + "; test -e " + rl.GzipPath() + " && cat " + rl.GzipPath() + " | gunzip; }"
}

func (rl *RemoteLog) Open() (reader io.ReadCloser, e error) {
	c := rl.Command()
	var cmd *exec.Cmd
	if rl.User == "" {
		rl.User = "root"
	}
	if rl.Host != "" {
		cmd = exec.Command("ssh", "-t", "-l", rl.User, rl.Host, c)
	} else {
		cmd = exec.Command("bash", "-c", c)
	}
	dbg.Printf("using cmd %q", cmd)
	reader, e = cmd.StdoutPipe()
	if e != nil {
		return nil, e
	}
	dbg.Print("starting command %q", c)
	e = cmd.Start()
	if e != nil {
		return nil, e
	}
	dbg.Print("command started")
	if rl.Compress {
		reader, e = gzip.NewReader(reader)
		if e != nil {
			return nil, e
		}
	}
	dbg.Print("returning reader")
	return reader, nil
}
