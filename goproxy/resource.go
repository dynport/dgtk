package goproxy

import (
	"net/http"
	"net/url"
	"os"
	"time"
)

var DefaultCacheDir = os.Getenv("HOME") + "/.goproxy/cache"

type Resource struct {
	Method string
	Url    *url.URL
	Header http.Header

	Response *Response
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
