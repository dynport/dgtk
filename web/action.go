package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Action interface {
	Template() ([]byte, error)
}

type LoadAction interface {
	Action
	Load(r *http.Request) error
}

type LoadWithParamsAction interface {
	Action
	Load(r *http.Request, params httprouter.Params) error
}
