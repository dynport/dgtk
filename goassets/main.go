package main

import (
	"flag"
	"github.com/dynport/dgtk/log"
	"os"
)

const BYTE_LENGTH = 12

func makeLineBuffer() []string {
	return make([]string, 0, BYTE_LENGTH)
}

var packageName = flag.String("pkg", "assets", "Package name to be used")
var packagePath = flag.String("path", "./assets.go", "Path to store assets.go")

func main() {
	flag.Parse()
	if len(os.Args) < 2 {
		log.Fatal("no asset path provided")
	}
	assets := &Assets{
		Package:           *packageName,
		CustomPackagePath: *packagePath,
		Paths:             flag.Args(),
	}
	if e := assets.Build(); e != nil {
		log.Fatal(e.Error())
	}
}
