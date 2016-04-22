package main

import (
	"log"
	"os"

	"github.com/dynport/dgtk/cli"
)

var logger = log.New(os.Stderr, "", 0)

func main() {
	router := cli.NewRouter()
	router.Register("aliases/create", &aliasCreate{}, "Create alias")
	router.Register("aliases/ls", &esAliases{}, "List Aliases")
	router.Register("aliases/rm", &aliasDelete{}, "Delete alias")
	router.Register("aliases/swap", &swapIndex{}, "Swap Alias")
	router.Register("index/dump", &dump{}, "Dump an index")
	router.Register("index/ls", &esIndexes{}, "List es indexes")
	router.Register("index/rm", &indexDelete{}, "Delete index")
	router.Register("index/stats", &indexStats{}, "Index Stats")
	router.Register("nodes/ls", &nodesLS{}, "Nodes List")
	router.Register("spy", &spy{}, "Spy on es requests")

	switch e := router.RunWithArgs(); e {
	case nil, cli.ErrorHelpRequested, cli.ErrorNoRoute:
		// ignore
		return
	default:
		logger.Fatal(e)
	}
}
