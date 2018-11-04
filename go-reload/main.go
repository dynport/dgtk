package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func main() {
	l := logrus.New()
	if err := run(l); err != nil {
		l.Fatalf("%+v", err)
	}
}

func run(l logrus.FieldLogger) error {
	path, err := exec.LookPath(os.Args[1])
	if err != nil {
		return errors.WithStack(err)
	}
	if len(os.Args) < 2 {
		return errors.New("at least 2 parameters required")
	}
	options := []string{}
	if len(os.Args) > 2 {
		options = os.Args[2:]
	}

	for {
		err := func() error {
			w, err := fsnotify.NewWatcher()
			if err != nil {
				return errors.WithStack(err)
			}
			defer w.Close()
			err = w.Add(path)
			if err != nil {
				return errors.WithStack(err)
			}
			l.Printf("starting up")
			ctx, cf := context.WithCancel(context.Background())
			go func() {
				_ = <-w.Events
				cf()
			}()
			c := exec.CommandContext(ctx, path, options...)
			c.Stdout = os.Stdout
			c.Stdin = os.Stdin
			c.Stderr = os.Stderr
			return c.Run()
		}()
		if err != nil {
			if !strings.Contains(err.Error(), "signal: killed") {
				l.Warnf("%T %v", err, err)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil
}
