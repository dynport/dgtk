package main

import (
	"log"
	"os"
)

func init() {
	if os.Getenv("DEBUG") == "true" {
		log.SetFlags(log.Llongfile)
	} else {
		log.SetFlags(0)
	}
}
