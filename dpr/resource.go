package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

func NewResource(dataRoot string, r *http.Request) *Resource {
	return &Resource{dataRoot: dataRoot, Request: r}
}

type Resource struct {
	dataRoot string
	*http.Request
}

func (r *Resource) root() string {
	if r.dataRoot == "" {
		panic("dataRoot must be set")
	}
	return r.dataRoot
}

func (r *Resource) localPath() string {
	normalizedPath := r.root() + r.Request.URL.Path
	if strings.HasSuffix(normalizedPath, "/") {
		normalizedPath += "index"
	}
	return normalizedPath
}

func (r *Resource) localHeadersPath() string {
	return r.localPath() + ".headers"
}

func (r *Resource) store() error {
	if e := r.storeHeaders(); e != nil {
		return e
	}
	if e := r.storeBody(); e != nil {
		return e
	}
	return nil
}

func (r *Resource) storeBody() error {
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

func (r *Resource) createFile(p string) (io.WriteCloser, error) {
	if e := os.MkdirAll(path.Dir(p), 0755); e != nil {
		return nil, e
	}
	log.Printf("creating file %s", p)
	return os.Create(p)
}

func (r *Resource) storeHeaders() error {
	f, e := r.createFile(r.localHeadersPath())
	if e != nil {
		return e
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(r.Header)
}

func (r *Resource) Write(w http.ResponseWriter) (int64, error) {
	log.Println("sending " + r.localPath())
	f, e := os.Open(r.localPath())
	if e != nil {
		return 0, e
	}
	defer f.Close()
	stat, e := f.Stat()
	if e != nil {
		return 0, e
	}
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	if strings.HasSuffix(r.Request.URL.String(), "/json") {
		stat, e := os.Stat(path.Dir(r.localPath()) + "/layer")
		if e == nil {
			w.Header().Add("X-Docker-Size", strconv.FormatInt(stat.Size(), 10))
		}
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(200)
	i, e := io.Copy(w, f)
	log.Printf("sent %d byte", i)
	return i, e
}

func (r *Resource) Exists() bool {
	return fileExists(r.localPath())
}
