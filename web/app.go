package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

func New(layout Layout) *App {
	return &App{
		Funcs:  template.FuncMap{},
		Layout: layout,
	}
}

type App struct {
	Layout       Layout
	DefaultTitle string
	Funcs        template.FuncMap

	urlFuncs map[string]UrlFunc
}

func (a *App) Url(name string, params ...interface{}) string {
	if a.urlFuncs == nil {
		panic("urlFuncs not set")
	}
	return a.urlFuncs[name](params...)
}

var logger = log.New(os.Stderr, "", 0)

func DefaultHandler(h func(w http.ResponseWriter, r *http.Request, params httprouter.Params)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h(w, r, nil)
	}
}

func (t *App) Handler(action Action) func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return t.LayoutHandler(t.Layout, action)
}

func (t *App) ActionHandler(action Action) func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		action, ok := clone(action).(Action)
		if !ok {
			t.HandleError(w, fmt.Errorf("type of action must be web.Action, was %T", action))
			return
		} else {
			t.HandleAction(w, r, action)
		}
	}
}

func renderAction(r *http.Request, action Action, params httprouter.Params, funcs template.FuncMap) ([]byte, error) {
	switch a := action.(type) {
	case LoadAction:
		return renderLoadAction(r, a, funcs)
	case LoadWithParamsAction:
		return renderLoadWithParamsAction(r, a, params, funcs)
	default:
		b, e := action.Template()
		if e != nil {
			return b, e
		}
		return render(b, funcs, action)
	}

}

func renderLoadWithParamsAction(r *http.Request, action LoadWithParamsAction, params httprouter.Params, funcs template.FuncMap) ([]byte, error) {
	b, e := action.Template()
	if e != nil {
		return nil, e
	}
	e = action.Load(r, params)
	if e != nil {
		return nil, e
	}
	return render(b, funcs, action)

}

func renderLoadAction(r *http.Request, action LoadAction, funcs template.FuncMap) ([]byte, error) {
	b, e := action.Template()
	if e != nil {
		return nil, e
	}
	switch a := action.(type) {
	case LoadAction:
		e = a.Load(r)
		if e != nil {
			return nil, e
		}
	}
	return render(b, funcs, action)
}

func (t *App) HandleAction(w http.ResponseWriter, r *http.Request, action Action) {
	var b []byte
	var e error
	switch a := action.(type) {
	case LoadAction:
		b, e = renderLoadAction(r, a, t.Funcs)
	}
	if e != nil {
		t.HandleError(w, e)
	} else {
		w.Write(b)
	}
}

// allow registering error pages
func (t *App) HandleError(w http.ResponseWriter, e error) {
	status := 500
	if s, ok := e.(interface {
		Status() int
	}); ok {
		status = s.Status()
	}
	logger.Printf("ERROR: %q", e)
	http.Error(w, e.Error(), status)
}
