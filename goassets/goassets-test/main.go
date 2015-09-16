package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	fs := FileSystem("")
	for _, n := range fs.AssetNames() {
		b, err := readAssetNew(fs, n)
		if err != nil {
			return err
		}
		fmt.Printf("%s: %d\n", n, len(b))
	}
	return nil
}

func readAssetNew(fs AssetFileSystem, name string) ([]byte, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}
