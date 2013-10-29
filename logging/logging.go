package logging

import (
	"bufio"
	"os/exec"
	"strings"
	"time"
)

type RemoteLog struct {
	Host    string
	Pattern string
	Tail    bool
	Time    time.Time
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

func (log *RemoteLog) Path() string {
	if !log.Time.IsZero() {
		return log.Time.Format("cat " + LOG_ROOT + "/" + HOURLY_PATTERN)
	}
	return CURRENT_ROOT
}

func (log *RemoteLog) GzipPath() string {
	return log.Path() + ".gz"
}

func (log *RemoteLog) Command() string {
	cmd := log.CatCmd()
	if log.Pattern != "" {
		cmd += " | " + log.GrepCmd()
	}
	return cmd
}

func (log *RemoteLog) GrepCmd() string {
	return "grep " + log.Pattern
}

func (log *RemoteLog) CatCmd() string {
	if log.Tail {
		return "tail -n 0 -F " + CURRENT_ROOT
	}
	return "{ test -e " + log.Path() + " && cat " + log.Path() + "; test -e " + log.GzipPath() + " && cat " + log.GzipPath() + " | gunzip; }"
}

func (log *RemoteLog) Each(pattern string, f func(line string)) error {
	cmd := exec.Command("ssh", "-l", "root", log.Host, log.Command())
	reader, e := cmd.StdoutPipe()
	if e != nil {
		return e
	}
	defer reader.Close()
	cmd.Start()
	scanner := bufio.NewScanner(reader)
	i := 0
	for scanner.Scan() {
		i++
		l := scanner.Text()
		if pattern != "" {
			if !strings.Contains(l, pattern) {
				continue
			}
		}
		f(l)
	}
	return nil
}
