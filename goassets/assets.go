package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Assets struct {
	Package string
	Assets  []*Asset
	Path    string
}

func (assets *Assets) Bytes() (b []byte, e error) {
	tpl := template.Must(template.New("assets").Parse(TPL))
	buf := &bytes.Buffer{}
	e = tpl.Execute(buf, assets)
	if e != nil {
		return b, e
	}
	return buf.Bytes(), nil
}

func (assets *Assets) Build() error {
	paths, e := filepath.Glob(assets.Path + "/*.*")
	if e != nil {
		return nil
	}
	for _, path := range paths {
		log.Print("adding " + path)
		if strings.HasSuffix(path, ".go") {
			continue
		}
		asset := &Asset{
			Path: path,
		}
		if e := asset.Load(); e != nil {
			log.Fatal(e.Error())
		}
		assets.Assets = append(assets.Assets, asset)
	}
	f, e := os.Create(assets.Path + "/assets.go")
	if e != nil {
		return e
	}
	defer f.Close()
	b, e := assets.Bytes()
	if e != nil {
		return e
	}
	_, e = f.Write(b)
	return e
}
