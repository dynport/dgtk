package goproxy

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type handler struct {
	Proxy *Proxy
}

func (handler *handler) DoStore(r *Resource) bool {
	if r.Response.StatuCode < 200 || r.Response.StatuCode >= 500 {
		return false
	}
	return !handler.Proxy.Ignored(handler.Proxy.cachePath(r))
}

func (handler *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	loaded, e := handler.Proxy.Load(res)
	logLine := fmt.Sprintf("method=%s url=%q size=%d", res.Method, res.Url.String(), len(res.Response.Body))
	if loaded {
		logLine += " status=loaded"

		if handler.DoStore(res) {
			e := handler.Proxy.Store(res)
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
