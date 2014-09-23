package main

import (
	"log"
	"os"

	"github.com/dynport/gocli"
)

var logger = log.New(os.Stderr, "", 0)

func main() {
	if e := run(); e != nil {
		logger.Fatal(e)
	}
}

func run() error {
	logger.Printf(gocli.Red("running"))
	return nil
}
