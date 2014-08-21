package web

import "net/http"

type Action interface {
	Load(r *http.Request) error
	Template() ([]byte, error)
}
