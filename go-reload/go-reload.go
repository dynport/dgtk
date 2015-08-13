package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/go-errors/errors"
)

var logger = log.New(os.Stderr, "[go-reload] ", log.Ldate|log.Ltime)

func main() {
	if err := run(); err != nil {
		logger.Fatal(errorMessage(err))
	}
}

func errorMessage(err error) string {
	switch e := err.(type) {
	case *errors.Error:
		return e.Error() + "\n" + e.ErrorStack()
	default:
		return e.Error()
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("at least 2 parameters required")
	}
	path, e := exec.LookPath(os.Args[1])
	if e != nil {
		return e
	}
	for {
		stat, e := os.Stat(path)
		if e != nil {
			return e
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
				return c.Start()
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
			return errors.New(err)
		}
		c.Wait()
	}
}
