package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Resource interface {
	Store() error
	Tags() (map[string]string, error)
	Exists() bool
	Open() (io.Reader, int64, error)
	DockerSize() (int64, error)
	LoadResource(p string) ([]byte, error)
}

func NewResource(dataRoot string, r *http.Request) Resource {
	if false {
		return &S3Resource{Request: r}
	}
	return &FileResource{dataRoot: dataRoot, Request: r}
}

type FileResource struct {
	dataRoot string
	*http.Request
}

func (r *FileResource) LoadResource(p string) ([]byte, error) {
	return ioutil.ReadFile(r.dataRoot + p)
}

func (r *FileResource) root() string {
	if r.dataRoot == "" {
		panic("dataRoot must be set")
	}
	return r.dataRoot
}

func (r *FileResource) localPath() string {
	normalizedPath := r.root() + r.Request.URL.Path
	if strings.HasSuffix(normalizedPath, "/") {
		normalizedPath += "index"
	}
	return normalizedPath
}

func (r *FileResource) Store() error {
	return r.storeBody()
}

func (r *FileResource) Tags() (map[string]string, error) {
	files, e := filepath.Glob(r.localPath() + "/*")
	if e != nil {
		return nil, e
	}
	tags := map[string]string{}
	for _, f := range files {
		b, e := ioutil.ReadFile(f)
		if e != nil {
			continue
		}
		tags[path.Base(f)] = strings.Replace(string(b), `"`, "", -1)
	}
	return tags, nil
}

func (r *FileResource) storeBody() error {
	f, e := r.createFile(r.localPath())
	if e != nil {
		return e
	}
	defer f.Close()
	i, e := io.Copy(f, r.Body)
	if e != nil {
		return e
	}
	log.Printf("stored %d byte", i)
	return nil
}

func (r *FileResource) createFile(p string) (io.WriteCloser, error) {
	if e := os.MkdirAll(path.Dir(p), 0755); e != nil {
		return nil, e
	}
	log.Printf("creating file %s", p)
	return os.Create(p)
}

func (r *FileResource) Open() (io.Reader, int64, error) {
	f, e := os.Open(r.localPath())
	if e != nil {
		return nil, 0, e
	}
	defer f.Close()
	stat, e := f.Stat()
	if e != nil {
		return nil, 0, e
	}
	return f, stat.Size(), nil
}

func (r *FileResource) DockerSize() (int64, error) {
	stat, e := os.Stat(path.Dir(r.localPath()) + "/layer")
	return stat.Size(), e
}

func (r *FileResource) Exists() bool {
	return fileExists(r.localPath())
}
