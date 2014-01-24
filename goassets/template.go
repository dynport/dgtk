package main

const TPL = `package {{ .Package }}

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

var assets = map[string][]byte{}

func assetNames() (names []string) {
	for name, _ := range assets {
		names = append(names, name)
	}
	return names
}

var (
	devAssetsPath string
	Development bool
	Debug bool
)

func init() {
	devAssetsPath = os.Getenv("DEV_ASSETS_PATH")
	if devAssetsPath != "" {
		Development = true
	}
}

func logDebug(format string, args ...interface{}) {
	if Debug {
		fmt.Printf(format + "\n", args...)
	}
}

func mustReadAsset(key string) []byte {
	content, e := readAsset(key)
	if e != nil {
		panic(e)
	}
	return content
}

func readAsset(key string) ([]byte, error) {
	if devAssetsPath != "" {
		path := devAssetsPath + "/" + key
		logDebug("reading file from dev path %s", path)
		b, e := ioutil.ReadFile(path)
		if e == nil {
			return b, nil
		}
	}
	b, ok := assets[key]
	if !ok {
		return nil, fmt.Errorf("asset %s not found in %v", key, assetNames())
	}
	gz, err := gzip.NewReader(bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("Decompression failed: %s", err.Error())
	}

	var buf bytes.Buffer
	io.Copy(&buf, gz)
	gz.Close()
	return buf.Bytes(), nil
}

func init() {
	{{ range .Assets }}assets["{{ .Key }}"] = []byte{
{{ .Bytes }}
	}
	{{ end }}
}
`
