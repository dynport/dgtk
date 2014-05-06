package goassets

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

const BYTE_LENGTH = 12

type Asset struct {
	Path  string
	Key   string
	Name  string
	Bytes string
}

func (asset *Asset) Load() error {
	buf := &bytes.Buffer{}
	gz := gzip.NewWriter(buf)
	f, e := os.Open(asset.Path)
	if e != nil {
		return e
	}
	defer f.Close()
	_, e = io.Copy(gz, f)
	gz.Flush()
	gz.Close()
	if e != nil {
		return e
	}
	list := make([]string, 0, len(buf.Bytes()))
	for _, b := range buf.Bytes() {
		list = append(list, fmt.Sprintf("0x%x", b))
	}
	buffer := makeLineBuffer()
	asset.Name = path.Base(asset.Path)
	for _, b := range list {
		buffer = append(buffer, b)
		if len(buffer) == BYTE_LENGTH {
			asset.Bytes += strings.Join(buffer, ",") + ",\n"
			buffer = makeLineBuffer()
		}
	}
	if len(buffer) > 0 {
		asset.Bytes += strings.Join(buffer, ",") + ",\n"
	}
	return nil
}

var debugger = log.New(debugStream(), "", 0)

func debugStream() io.Writer {
	if os.Getenv("DEBUG") == "true" {
		return os.Stderr
	}
	return ioutil.Discard
}

func (assets *Assets) build() ([]byte, error) {
	if assets.Package == "" {
		assets.Package = "main"
	}
	debugger.Print("loading assets paths")
	paths, e := assets.AssetPaths()
	debugger.Printf("got %d assets", len(paths))
	if e != nil {
		return nil, e
	}
	for _, asset := range paths {
		debugger.Printf("loading assets %q", asset.Key)
		e := asset.Load()
		if e != nil {
			return nil, e
		}
		assets.Assets = append(assets.Assets, asset)
	}
	return assets.Bytes()
}

func (assets *Assets) Build() error {
	b, e := assets.build()
	if e != nil {
		return e
	}
	path, e := assets.PackagePath()
	if e != nil {
		return e
	}
	if fileExists(path) {
		return fmt.Errorf("File %q already exists (deleted it first?!?)", path)
	}
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	_, e = f.Write(b)
	return e
}

func makeLineBuffer() []string {
	return make([]string, 0, BYTE_LENGTH)
}
