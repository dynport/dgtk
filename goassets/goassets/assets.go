package goassets

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type Assets struct {
	Package           string
	CustomPackagePath string
	Assets            []*Asset
	Paths             []string
	BuiltAt           string
}

func (assets *Assets) Bytes() (b []byte, e error) {
	tpl := template.Must(template.New("assets").Parse(TPL))
	buf := &bytes.Buffer{}
	assets.BuiltAt = time.Now().UTC().Format(time.RFC3339Nano)
	e = tpl.Execute(buf, assets)
	if e != nil {
		return b, e
	}
	return buf.Bytes(), nil
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
	e = filepath.Walk(path, func(p string, stat os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if stat.IsDir() {
			return nil
		}
		abs, e := filepath.Abs(p)
		if e != nil {
			return e
		}
		if abs != packagePath {
			assets = append(assets, &Asset{Path: p})
		}
		return nil
	})
	return assets, e
}

func (assets *Assets) PackagePath() (path string, e error) {
	path = assets.CustomPackagePath
	if path == "" {
		path = "./assets.go"
	}
	return filepath.Abs(path)
}
