package goproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

func New(root string) (*Proxy, error) {
	db, e := leveldb.OpenFile(root, nil)
	if e != nil {
		return nil, e
	}
	return &Proxy{db: db}, nil
}

func (proxy *Proxy) Ignored(p string) bool {
	_, ok := proxy.ignored[path.Base(p)]
	return ok
}

func (proxy *Proxy) Ignore(name string) {
	if proxy.ignored == nil {
		proxy.ignored = map[string]struct{}{}
	}
	proxy.ignored[name] = struct{}{}
}

func (proxy *Proxy) Run() error {
	db, e := leveldb.OpenFile(proxy.CacheDir(), nil)
	if e != nil {
		return e
	}
	proxy.db = db
	handler := &handler{Proxy: proxy}
	return http.ListenAndServe(proxy.Address, handler)
}

type Proxy struct {
	Address        string
	CustomCacheDir string
	db             *leveldb.DB
	ignored        map[string]struct{}
}

func (proxy *Proxy) Store(r *Resource) error {
	buf := &bytes.Buffer{}
	e := json.NewEncoder(buf).Encode(r)
	if e != nil {
		return e
	}
	return proxy.db.Put(proxy.cacheKey(r), buf.Bytes(), nil)
}

func (proxy *Proxy) cacheKey(r *Resource) []byte {
	return []byte("resources/" + r.Url.RequestURI())
}

func (proxy *Proxy) loadCached(r *Resource) (loaded bool, e error) {
	b, e := proxy.db.Get(proxy.cacheKey(r), nil)
	if e != nil {
		return false, e
	}
	e = json.Unmarshal(b, r)
	return true, e
}

func (proxy *Proxy) Load(r *Resource) (bool, error) {
	m := NewMessage("resources.store", r)
	m.publish("started")
	defer m.publish("finished")
	if !r.validMethod() {
		m.Error = fmt.Errorf("%s is not a valid method", r.Method)
		return false, m.Error
	}
	if proxy.cached(r) {
		loaded, e := proxy.loadCached(r)
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

func (proxy *Proxy) cached(r *Resource) bool {
	it := proxy.db.NewIterator(nil, nil)
	key := proxy.cacheKey(r)
	return it.Seek(key)
}

func (proxy *Proxy) CacheDir() string {
	if proxy.CustomCacheDir != "" {
		return proxy.CustomCacheDir
	}
	return DefaultCacheDir
}

func (proxy *Proxy) cachePath(r *Resource) string {
	dir := proxy.CacheDir() + "/" + r.Url.Host
	p := r.Url.Path
	switch p {
	case "", "/":
		p = "/__index"
	}
	if strings.HasSuffix(r.Url.Path, "/") {
		p += "__index"
	}
	if query := r.Url.Query().Encode(); query != "" {
		p += "/" + query
	}
	return dir + p
}
