package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
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
	if len(os.Args) < 2 {
		return errors.New("at least 2 parameters required")
	}
	path, err := exec.LookPath(os.Args[1])
	if err != nil {
		return errors.WithStack(err)
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.WithStack(err)
	}
	err = w.Add(path)
	if err != nil {
		return errors.WithStack(err)
	}

	options := []string{}
	if len(os.Args) > 2 {
		options = os.Args[2:]
	}

	var (
		ctx context.Context
		cf  func()
		mux sync.Mutex
	)

	go func() {
		for {
			l.Printf("starting up")
			mux.Lock()
			ctx, cf = context.WithCancel(context.Background())
			mux.Unlock()
			c := exec.CommandContext(ctx, path, options...)
			c.Stdout = os.Stdout
			c.Stdin = os.Stdin
			c.Stderr = os.Stderr
			err = c.Run()
			if err != nil {
				if !strings.Contains(err.Error(), "signal: killed") {
					l.Warnf("%T %v", err, err)
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	for _ = range w.Events {
		mux.Lock()
		cf()
		mux.Unlock()
	}
	return nil
}
