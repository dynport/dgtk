package main

import (
	"fmt"
	"os"
)

var debug = os.Getenv("DEBUG") == "true"

func log(format string, i ...interface{}) {
	if debug {
		fmt.Printf(format+"\n", i...)
	}
}

func logFatal(format string, i ...interface{}) {
	debug = true
	log("ERROR: "+format, i...)
	os.Exit(1)
}
