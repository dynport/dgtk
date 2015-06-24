package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func run() error {
	args := []string{"test"}
	if len(os.Args) > 1 {
		args = append(args, os.Args[1:]...)
	}
	c := exec.Command("go", args...)
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	c.Env = append(os.Environ(), "TEST_ROOT="+wd)
	stdErr, err := c.StderrPipe()
	if err != nil {
		return err
	}
	stdOut, err := c.StdoutPipe()
	if err != nil {
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go workStream(stdOut, wg)
	go workStream(stdErr, wg)

	if err := c.Start(); err != nil {
		return err
	}
	if err := c.Wait(); err != nil {
		return err
	}
	wg.Wait()
	return nil
}

func workStream(in io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\r")
		fmt.Println(parts[len(parts)-1])
	}
}
