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

func Names() (names []string) {
	for name, _ := range assets {
		names = append(names, name)
	}
	return names
}

var (
	devAssetsPath string
	Development bool
)

func init() {
	devAssetsPath = os.Getenv("DEV_ASSETS_PATH")
	if devAssetsPath != "" {
		Development = true
	}
}

func Get(key string) ([]byte, error) {
	if devAssetsPath != "" {
		return ioutil.ReadFile(devAssetsPath + "/" + key)
	}
	b, ok := assets[key]
	if !ok {
		return nil, fmt.Errorf("asset %s not found in %v", key, Names())
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
	{{ range .Assets }}assets["{{ .Name }}"] = []byte{
{{ .Bytes }}
	}
	{{ end }}
}
`
