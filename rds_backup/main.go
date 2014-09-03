package main

import (
	"log"

	"github.com/dynport/dgtk/cli"
)

func main() {
	switch e := router().RunWithArgs(); e {
	case nil, cli.ErrorNoRoute, cli.ErrorHelpRequested:
		// ignore

	default:
		log.Fatal(e)
	}
}
