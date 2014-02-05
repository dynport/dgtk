package main

const TPL = `package {{ .Package }}

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type assetFileSystemI interface {
	Open(name string) (assetFileI, error)
	AssetNames() []string
}

type assetFileI interface {
	io.Closer
	io.Reader
}

var assets assetFileSystemI

func assetNames() (names []string) {
	return assets.AssetNames()
}

func readAsset(key string) ([]byte, error) {
	r, e := assets.Open(key)
	if e != nil {
		return nil, e
	}
	defer r.Close()

	p, e := ioutil.ReadAll(r)
	if e != nil {
		return nil, e
	}
	return p, nil
}

func mustReadAsset(key string) []byte {
	p, e := readAsset(key)
	if e != nil {
		panic(e)
	}
	return p
}

type assetOsFS struct{ root string }

func (aFS assetOsFS) Open(name string) (assetFileI, error) {
	return os.Open(filepath.Join(aFS.root, name))
}

func (aFS *assetOsFS) AssetNames() ([]string) {
	names, e := filepath.Glob(aFS.root + "/*")
	if e != nil {
		log.Print(e)
	}
	return names
}

type assetIntFS map[string][]byte

type assetFile struct {
	decompressor io.ReadCloser
}

func (afs assetIntFS) AssetNames() (names []string) {
	names = make([]string, 0, len(afs))
	for k, _ := range afs {
		names = append(names, k)
	}
	return names
}

func (afs assetIntFS) Open(name string) (af assetFileI, e error) {
	if asset, found := afs[name]; found {
		decomp, e := gzip.NewReader(bytes.NewBuffer(asset))
		if e != nil {
			return nil, e
		}
		af = &assetFile{decompressor: decomp}
		return af, nil
	}
	return nil, os.ErrNotExist
}

func (a *assetFile) Close() error {
	if a.decompressor != nil {
		return a.decompressor.Close()
	}
	return nil
}

func (a *assetFile) Read(p []byte) (n int, e error) {
	if a.decompressor == nil {
		return 0, os.ErrInvalid
	}
	return a.decompressor.Read(p)
}

func init() {
	env_name := fmt.Sprintf("GOASSETS_%s_PATH", strings.ToUpper("{{ .Package }}"))
	path := os.Getenv(env_name)
	if path != "" {
		if !filepath.IsAbs(path) {
			log.Fatalf("path %q given in %s must be absolute!", path, env_name)
		}
		if _, e := os.Stat(path); e != nil {
			log.Fatalf("path %q does not exist!", path)
		}
		assets = &assetOsFS{root: path}
		return
	}

	assetsTmp := assetIntFS{}
	{{ range .Assets }}assetsTmp["{{ .Key }}"] = []byte{
		{{ .Bytes }}
	}
	{{ end }}
	assets = assetsTmp
}
`
