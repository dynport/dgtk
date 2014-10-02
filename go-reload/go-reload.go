package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

var logger = log.New(os.Stderr, "[go-reload] ", log.Ldate|log.Ltime)

func main() {
	e := run()
	if e != nil {
		logger.Fatal(e)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("at least 2 parameters required")
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
		c := exec.Command(path, options...)
		c.Stdout = os.Stdout
		c.Stdin = os.Stdin
		c.Stderr = os.Stderr
		e = c.Start()
		if e != nil {
			return e
		}
		logger.Printf("running with pid %d", c.Process.Pid)
		for {
			stat, e := os.Stat(path)
			if e != nil {
				logger.Printf("ERROR: %s", e)
			} else if stat.ModTime() != modified {
				logger.Printf("mod time changed => sleeping")
				break
			}
			time.Sleep(1 * time.Second)
		}
		logger.Printf("killing pid %d", c.Process.Pid)
		e = c.Process.Kill()
		if e != nil {
			return e
		}
	}
}
