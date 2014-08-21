package web

import (
	"html/template"
	"net/http"
)

type Layout interface {
	Load(*http.Request, Action, template.HTML) error
	Template() ([]byte, error)
}
