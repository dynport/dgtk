package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	builtAt        = time.Now()
	compiledAssets = assetIntFS{}
	assets         AssetFileSystem
)

func debugStream() io.Writer {
	if os.Getenv("DEBUG") == "true" {
		return os.Stderr
	}
	return ioutil.Discard
}

var dbg = log.New(debugStream(), "[DEBUG] ", log.Lshortfile)

type assetProxy struct {
	devPath string
}

func (a *assetProxy) Open(name string) (http.File, error) {
	return a.fileSystem().Open(name)
}

func (a *assetProxy) AssetNames() []string {
	return a.fileSystem().AssetNames()
}

func (a *assetProxy) fileSystem() AssetFileSystem {
	dbg.Printf("getting file system for %q", a.devPath)
	if a.devPath != "" {
		dbg.Printf("using dev path %s", a.devPath)
		stat, e := os.Stat(a.devPath)
		if e == nil && stat.IsDir() {
			assets = &assetOsFS{root: a.devPath}
			return assets
		} else {
			dbg.Printf("dev path %s does not exist", a.devPath)
		}
	} else {
		dbg.Printf("dev path seems to be empty")
	}
	return compiledAssets
}

func FileSystem(devPath string) AssetFileSystem {
	return &assetProxy{devPath: devPath}
}

type AssetFileSystem interface {
	Open(name string) (http.File, error)
	AssetNames() []string
}

type assetOsFS struct{ root string }

func (aFS assetOsFS) Open(name string) (http.File, error) {
	p := filepath.Join(aFS.root, name)
	dbg.Printf("opening local file %q", p)
	f, e := os.Open(p)
	if e != nil {
		dbg.Printf("ERROR reading local file: %q", name)
		return nil, e
	}
	return f, nil
}

func (aFS *assetOsFS) AssetNames() []string {
	names, e := filepath.Glob(aFS.root + "/*")
	if e != nil {
		log.Print(e)
	}
	return names
}

type assetIntFS map[string][]byte

type assetNode struct {
	name string
	data *bytes.Reader
	dir  bool

	children map[string]*assetNode
}

func addNode(root *assetNode, path string, content *bytes.Reader) error {
	node := root
	pathSegments := strings.Split(path, "/")
	if len(pathSegments) > 1 {
		for i := 0; i < len(pathSegments)-1; i++ {
			if val, ok := node.children[pathSegments[i]]; ok {
				node = val
			} else {
				newNode := &assetNode{name: pathSegments[i], dir: true, children: map[string]*assetNode{}}
				node.children[pathSegments[i]] = newNode
				node = newNode
			}
		}
	}
	filename := pathSegments[len(pathSegments)-1]
	if _, ok := node.children[filename]; ok {
		return fmt.Errorf("node %q already exists", filename)
	}
	node.children[filename] = &assetNode{name: filename, data: content}
	return nil
}

func (node *assetNode) Traverse(path []string) (*assetNode, error) {
	switch len(path) {
	case 0:
		return node, nil
	default:
		child, ok := node.children[path[0]]
		if !ok {
			return nil, os.ErrNotExist
		}
		return child.Traverse(path[1:])
	}
}

func (node *assetNode) Name() string {
	return node.name
}

func (node *assetNode) ModTime() time.Time {
	return builtAt
}

func (node *assetNode) Mode() os.FileMode {
	if node.dir {
		return 0755
	}
	return 0644
}

func (node *assetNode) Sys() interface{} {
	return nil
}

func (node *assetNode) Size() int64 {
	if node.dir {
		return 0
	}
	return int64(node.data.Len())
}

func (node *assetNode) IsDir() bool {
	return node.dir
}

func (node *assetNode) Readdir(count int) (stats []os.FileInfo, e error) {
	if !node.dir {
		return nil, nil
	}

	for _, child := range node.children {
		stat, e := child.Stat()
		if e != nil {
			return nil, e
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

func (node *assetNode) Stat() (os.FileInfo, error) {
	return node, nil
}

func (node *assetNode) Close() error {
	return nil
}

func (node *assetNode) Read(p []byte) (int, error) {
	return node.data.Read(p)
}

func (node *assetNode) Seek(offset int64, whence int) (int64, error) {
	if node.dir {
		return 0, nil
	}
	return node.data.Seek(offset, whence)
}

func (node *assetNode) Open(name string) (af http.File, e error) {
	dbg.Printf("opening tpl %s", name)
	if name == "." {
		return node, nil
	}
	name = strings.TrimPrefix(name, "/")
	nameSegments := strings.Split(name, "/")
	return node.Traverse(nameSegments)
}

func (afs assetIntFS) AssetNames() (names []string) {
	names = make([]string, 0, len(afs))
	for k, _ := range afs {
		names = append(names, k)
	}
	return names
}

func (afs assetIntFS) Open(name string) (af http.File, e error) {
	dbg.Printf("opening tpl %s", name)

	switch name {
	case "":
		name = "index.html"
	case "/":
		name = ""
	default:
		name = strings.TrimPrefix(name, "/")
	}

	// single asset referenced, load it directly
	if asset, found := afs[name]; found {
		reader, e := createReader(asset)
		af = &assetNode{data: reader, name: name}
		return af, e
	}

	// directory request?
	switch {
	case name == "":
		// ignore
	case !strings.HasSuffix(name, "/"):
		name += "/"
	}
	root := &assetNode{dir: true, name: ".", children: map[string]*assetNode{}}
	for k, v := range afs {
		if strings.HasPrefix(k, name) {
			reader, e := createReader(v)
			if e != nil {
				return nil, e
			}
			dbg.Printf("adding node %q", k)
			addNode(root, strings.TrimPrefix(k, name), reader)
		}
	}
	if len(root.children) > 0 {
		return root, nil
	}

	dbg.Printf("ERROR: index %s does not exist. known keys: %#v", name, afs.AssetNames())
	return nil, os.ErrNotExist
}

func createReader(data []byte) (*bytes.Reader, error) {
	decomp, e := gzip.NewReader(bytes.NewBuffer(data))
	if e != nil {
		return nil, e
	}
	defer func() {
		_ = decomp.Close()
	}()
	b, e := ioutil.ReadAll(decomp)
	if e != nil {
		return nil, e
	}
	return bytes.NewReader(b), nil
}
