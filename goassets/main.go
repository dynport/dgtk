package main

import (
	"os"
)

const BYTE_LENGTH = 12

func makeLineBuffer() []string {
	return make([]string, 0, BYTE_LENGTH)
}

func main() {
	if len(os.Args) < 2 {
		logFatal("no asset path provided")
	}
	assets := &Assets{
		Package: "assets",
		Path:    os.Args[1],
	}
	if e := assets.Build(); e != nil {
		logFatal(e.Error())
	}
}
