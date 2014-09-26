package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/dynport/gocli"
)

func main() {
	if e := run(); e != nil {
		log.Fatal(e)
	}
}

var BUILD_INFO string

func run() error {
	b, e := base64.StdEncoding.DecodeString(BUILD_INFO)
	if e != nil {
		return e
	}
	fmt.Println(string(b))
	_ = gocli.Red("test") // just to add external dependencies
	return nil
}
