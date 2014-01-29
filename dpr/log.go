package main

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "", 0)

func init() {
	if os.Getenv("DEBUG") == "true" {
		logger.SetFlags(log.Llongfile)
		log.SetFlags(log.Llongfile)
	} else {
		log.SetFlags(0)
	}
}
