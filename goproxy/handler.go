package goproxy

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
)

type Handler struct {
	ignored map[string]struct{}
}

func (handler *Handler) Ignored(p string) bool {
	_, ok := handler.ignored[path.Base(p)]
	return ok
}

func (handler *Handler) DoStore(r *Resource) bool {
	if r.Response.StatuCode < 200 || r.Response.StatuCode >= 500 {
		return false
	}
	return !handler.Ignored(r.cachePath())
}

func (handler *Handler) Ignore(name string) {
	if handler.ignored == nil {
		handler.ignored = map[string]struct{}{}
	}
	handler.ignored[name] = struct{}{}
}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	u := r.URL.String()
	if strings.HasPrefix(u, "//") {
		u = "https:" + u
	}
	started := time.Now()
	res := &Resource{
		Method: r.Method,
		Url:    r.URL,
		Header: r.Header,
	}

	loaded, e := res.Load()
	logLine := fmt.Sprintf("method=%s url=%q size=%d", res.Method, res.Url.String(), len(res.Response.Body))
	if loaded {
		logLine += " status=loaded"

		if handler.DoStore(res) {
			e := res.store()
			if e != nil {
				log.Println("ERROR: " + e.Error())
			} else {
				logLine += " stored=true"
			}
		} else {
			logLine += " stored=ignored"
		}
	} else {
		logLine += " status=cached"
	}
	log.Print(logLine + fmt.Sprintf(" total_time=%.03f", time.Since(started).Seconds()))
	if e != nil {
		http.Error(w, e.Error(), 500)
		return
	}

	for k, values := range res.Response.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(res.Response.StatuCode)
	_, e = w.Write(res.Response.Body)
	if e != nil {
		log.Println("ERROR: " + e.Error())
	}
}
