package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/dynport/dgtk/cli"
)

var logger = log.New(os.Stderr, "", 0)

func debugStream() io.Writer {
	if os.Getenv("DEBUG") == "true" {
		return os.Stderr
	}
	return ioutil.Discard
}

var dbg = log.New(debugStream(), "[DEBUG] ", log.Lshortfile)

func main() {
	switch e := router().RunWithArgs(); e {
	case nil, cli.ErrorNoRoute, cli.ErrorHelpRequested:
		// ignore

	default:
		logger.Fatal(e)
	}
}
