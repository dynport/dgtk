package goproxy

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

var DefaultCacheDir = os.Getenv("HOME") + "/.goproxy/cache"

type Resource struct {
	Method string
	Url    *url.URL
	Header http.Header

	Response *Response

	CacheDir string
}

func (r *Resource) cachePath() string {
	dir := r.cacheDir() + "/" + r.Url.Host
	p := r.Url.Path
	switch p {
	case "", "/":
		p = "/index"
	}
	if query := r.Url.Query().Encode(); query != "" {
		p += "/" + query
	}
	return dir + p
}

func (r *Resource) cachePathMd5() string {
	cs := r.checksum()
	return r.cacheDir() + "/" + strings.Join(strings.Split(cs[0:4], ""), "/") + "/" + cs + ".json"
}

func (r *Resource) cached() bool {
	return fileExists(r.cachePath())
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func (r *Resource) cacheDir() string {
	if r.CacheDir != "" {
		return r.CacheDir
	}
	return DefaultCacheDir
}

func (r *Resource) checksum() string {
	payload := strings.Join(append([]string{r.Method, r.Url.String()}), "\n")

	sum := md5.New()
	sum.Write([]byte(payload))
	return fmt.Sprintf("%x", sum.Sum(nil))
}

func (r *Resource) loadCached() (loaded bool, e error) {
	f, e := os.Open(r.cachePath())
	if e != nil {
		return false, e
	}
	defer f.Close()
	e = json.NewDecoder(f).Decode(r)
	return true, e
}

func (r *Resource) Load() (fetched bool, e error) {
	m := NewMessage("resources.store", r)
	m.publish("started")
	defer m.publish("finished")
	if !r.validMethod() {
		m.Error = fmt.Errorf("%s is not a valid method", r.Method)
		return false, m.Error
	}
	if r.cached() {
		loaded, e := r.loadCached()
		if e == nil && loaded {
			m.Status = "cached"
			return false, nil
		} else {
			m.Error = e
			m.Status = "error_loading_cache"
			log.Printf("ERROR: %s", e.Error())
		}
	}
	m.Status = "loaded"
	req, e := http.NewRequest(r.Method, r.Url.String(), nil)
	if e != nil {
		m.publishError(e)
		return false, e
	}
	for k, values := range r.Header {
		for _, v := range values {
			req.Header.Add(k, v)
		}
	}
	rsp, e := http.DefaultClient.Do(req)
	if e != nil {
		return false, e
	}
	defer rsp.Body.Close()
	r.Response = &Response{
		Header:    http.Header{},
		FetchedAt: time.Now(),
	}
	r.Response.StatuCode = rsp.StatusCode
	m.publish("response")
	for k, values := range rsp.Header {
		r.Response.Header[k] = values
	}
	r.Response.Body, e = ioutil.ReadAll(rsp.Body)
	if e != nil {
		return false, e
	}
	m.publish("read")
	return true, e
}

func (r *Resource) store() error {
	p := r.cachePath()
	tmpPath := p + ".json"
	e := os.MkdirAll(path.Dir(tmpPath), 0755)
	if e != nil {
		return e
	}
	f, e := os.Create(tmpPath)
	if e != nil {
		return e
	}
	defer f.Close()
	e = json.NewEncoder(f).Encode(r)
	if e != nil {
		return e
	}
	return os.Rename(tmpPath, p)
}

func (r *Resource) validMethod() bool {
	switch r.Method {
	case "GET", "HEAD":
		return true
	}
	return false
}

type Response struct {
	Header    http.Header
	Body      []byte
	StatuCode int
	FetchedAt time.Time
}
