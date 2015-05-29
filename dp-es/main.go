package main

import (
	"log"
	"os"

	"github.com/dynport/dgtk/cli"
)

var logger = log.New(os.Stderr, "", 0)

func main() {
	router := cli.NewRouter()
	router.Register("indexes/list", &esIndexes{}, "List es indexes")
	router.Register("aliases/list", &esAliases{}, "List Aliases")
	router.Register("aliases/swap", &swapIndex{}, "Swap Alias")
	router.Register("aliases/create", &aliasCreate{}, "Create alias")
	router.Register("aliases/delete", &aliasDelete{}, "Delete alias")
	router.Register("indexes/delete", &indexDelete{}, "Delete index")
	switch e := router.RunWithArgs(); e {
	case nil, cli.ErrorHelpRequested, cli.ErrorNoRoute:
		// ignore
		return
	default:
		logger.Fatal(e)
	}
}
