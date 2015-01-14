package main

import (
	"log"

	"github.com/dynport/dgtk/cli"
)

func main() {
	e := cli.RunActionWithArgs(&sendNotification{})
	switch e {
	case nil, cli.ErrorNoRoute, cli.ErrorHelpRequested:
		// ignore
	default:
		log.Fatal(e)
	}
}
