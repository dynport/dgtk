package main

import (
	"bytes"
	"fmt"
	"github.com/dynport/dgtk/log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

type Assets struct {
	Package           string
	CustomPackagePath string
	Assets            []*Asset
	Paths             []string
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

func (assets *Assets) GetterMethodName() string {
	if assets.Package == "assets" {
		return "Get"
	} else {
		return "ReadAsset"
	}
}

func (assets *Assets) NamesMethodName() string {
	if assets.Package == "assets" {
		return "Names"
	} else {
		return "AssetNames"
	}
}

func (assets *Assets) AssetPaths() (out []*Asset, e error) {
	out = []*Asset{}
	packagePath, e := assets.PackagePath()
	if e != nil {
		return out, e
	}
	for _, path := range assets.Paths {
		tmp, e := assetsInPath(path, packagePath)
		if e != nil {
			return out, e
		}
		for _, asset := range tmp {
			asset.Key, e = removePrefix(asset.Path, path)
			if e != nil {
				return out, e
			}
			out = append(out, asset)
		}
	}
	return out, nil
}

func removePrefix(path, prefix string) (suffix string, e error) {
	absPath, e := filepath.Abs(path)
	if e != nil {
		return "", e
	}
	absPrefix, e := filepath.Abs(prefix)
	if e != nil {
		return "", e
	}
	if strings.HasPrefix(absPath, absPrefix) {
		return strings.TrimPrefix(strings.TrimPrefix(absPath, absPrefix), "/"), nil
	}
	return "", fmt.Errorf("%s has no prefix %s", absPath, absPrefix)
}

func assetsInPath(path string, packagePath string) (assets []*Asset, e error) {
	stat, e := os.Stat(path)
	if e != nil {
		return assets, e
	}
	if stat.IsDir() {
		paths, e := filepath.Glob(path + "/*")
		if e != nil {
			return assets, e
		}
		sort.Strings(paths)
		for _, path := range paths {
			stat, e := os.Stat(path)
			if e != nil {
				return assets, e
			}
			if stat.IsDir() {
				tmp, e := assetsInPath(path, packagePath)
				if e != nil {
					return assets, e
				}
				assets = append(assets, tmp...)
			} else {
				abs, e := filepath.Abs(path)
				if e != nil {
					return assets, e
				}
				if abs != packagePath {
					assets = append(assets, &Asset{Path: path})
				}
			}
		}
	}
	return assets, nil
}

func (assets *Assets) PackagePath() (path string, e error) {
	path = assets.CustomPackagePath
	if path == "" {
		path = "./assets.go"
	}
	return filepath.Abs(path)
}

func (assets *Assets) Build() error {
	paths, e := assets.AssetPaths()
	if e != nil {
		return e
	}
	for _, asset := range paths {
		log.Debug("loading %s", asset.Path)
		e := asset.Load()
		if e != nil {
			return e
		}
		assets.Assets = append(assets.Assets, asset)
	}
	path, e := assets.PackagePath()
	if e != nil {
		return e
	}
	if fileExists(path) {
		return fmt.Errorf("file %s already exists (deleted it first?!?)", path)
	}
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	b, e := assets.Bytes()
	if e != nil {
		return e
	}
	log.Debug("writing %d bytes to %s", len(b), path)
	_, e = f.Write(b)
	return e
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}
