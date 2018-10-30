package main

import (
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger = func() *logrus.Logger {
	l := logrus.New()
	return l
}()

func main() {
	if err := run(); err != nil {
		logger.Fatalf("%+v", err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("at least 2 parameters required")
	}
	path, err := exec.LookPath(os.Args[1])
	if err != nil {
		return errors.WithStack(err)
	}
	for {
		err := func() error {
			stat, err := os.Stat(path)
			if err != nil {
				return errors.WithStack(err)
			}
			modified := stat.ModTime()
			options := []string{}
			if len(os.Args) > 2 {
				options = os.Args[2:]
			}
			var c *exec.Cmd
			for i := 0; i < 10; i++ {
				err := func() error {
					c = exec.Command(path, options...)
					c.Stdout = os.Stdout
					c.Stdin = os.Stdin
					c.Stderr = os.Stderr
					return errors.WithStack(c.Start())
				}()
				if err != nil {
					time.Sleep(1 * time.Second)
				} else {
					break
				}
			}
			logger.Printf("running with pid %d", c.Process.Pid)
			for {
				if stat, err := os.Stat(path); err != nil {
					logger.Printf("ERROR: %s", err)
				} else if stat.ModTime() != modified {
					logger.Printf("mod time changed => sleeping")
					break
				}
				time.Sleep(1 * time.Second)
			}
			logger.Printf("killing pid %d", c.Process.Pid)
			if err := c.Process.Kill(); err != nil {
				return errors.WithStack(err)
			}
			c.Wait()
			return nil
		}()
		if err != nil {
			logger.Warn(err)
			time.Sleep(1 * time.Second)
		}
	}
}
