package main

import (
	"log"
	"os"

	"github.com/dynport/dgtk/cli"
)

var logger = log.New(os.Stderr, "", 0)

func main() {
	router := cli.NewRouter()
	router.Register("indexes/ls", &esIndexes{}, "List es indexes")
	router.Register("aliases/ls", &esAliases{}, "List Aliases")
	router.Register("aliases/swap", &swapIndex{}, "Swap Alias")
	router.Register("aliases/create", &aliasCreate{}, "Create alias")
	router.Register("aliases/rm", &aliasDelete{}, "Delete alias")
	router.Register("indexes/rm", &indexDelete{}, "Delete index")
	router.Register("spy", &spy{}, "Spy on es requests")

	switch e := router.RunWithArgs(); e {
	case nil, cli.ErrorHelpRequested, cli.ErrorNoRoute:
		// ignore
		return
	default:
		logger.Fatal(e)
	}
}
