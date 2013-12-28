package main

import (
	"github.com/dynport/dpgtk/cli2"
	_ "github.com/dynport/gossh"
	"log"
)

var router = cli2.NewRouter()

func must(e error) {
	if e != nil {
		panic(e.Error())
	}
}

func main() {
	e := router.RunWithArgs()
	if e != nil {
		log.Fatal(e.Error())
	}
}
