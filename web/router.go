package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type Router map[string]map[string]interface{}

func (r Router) UrlHelpers() (map[string]UrlFunc, error) {
	out := map[string]UrlFunc{}
	for _, routes := range r {
		for path, handle := range routes {
			if name := nameFor(handle); name != "" {
				if _, ok := out[name]; ok {
					return nil, fmt.Errorf("route for %q defined twice", name)
				}
				out[name] = parseUrlFunc(path)
			}
		}
	}
	return out, nil
}

func nameFor(handle interface{}) string {
	switch t := handle.(type) {
	case Action:
		parts := strings.Split(strings.TrimPrefix(fmt.Sprintf("%T", t), "*"), ".")
		return parts[len(parts)-1]
	case *NamedRoute:
		return t.Name
	case NamedRoute:
		return t.Name
	}
	return ""

}

func handleFor(app *App, i interface{}) httprouter.Handle {
	switch h := i.(type) {
	case Action:
		return app.Handler(h)
	case func(http.ResponseWriter, *http.Request, httprouter.Params):
		return h
	case func(http.ResponseWriter, *http.Request):
		return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) { h(w, r) }
	case *NamedRoute:
		return handleFor(app, h.Handler)
	case NamedRoute:
		return handleFor(app, h.Handler)
	}
	return nil
}

func (r Router) Router(app *App) (router *httprouter.Router, e error) {
	defer func() {
		if r := recover(); r != nil {
			e = fmt.Errorf("%s", r)
		}
	}()

	router = httprouter.New()

	for method, routes := range r {
		for path, i := range routes {
			handle := handleFor(app, i)
			if handle == nil {
				return nil, fmt.Errorf("no handler defined for %T", i)
			}
			switch method {
			case "PUT":
				router.PUT(path, handle)
			case "PATCH":
				router.PATCH(path, handle)
			case "DELETE":
				router.DELETE(path, handle)
			case "POST":
				router.POST(path, handle)
			case "GET":
				router.GET(path, handle)
			}
		}
	}

	if app.Funcs == nil {
		app.Funcs = template.FuncMap{}
	}
	if app.urlFuncs == nil {
		app.urlFuncs = map[string]UrlFunc{}
	}

	helpers, e := r.UrlHelpers()
	if e != nil {
		return nil, e
	}

	for n, f := range helpers {
		if _, ok := app.Funcs[n]; ok {
			return nil, fmt.Errorf("helper %q defined twice", n)
		}
		app.Funcs[n+"Path"] = f
		app.urlFuncs[n] = f
	}
	return router, e

}

func h(app *App, action Action) func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		app.Handler(action)(w, r, params)
	}
}
