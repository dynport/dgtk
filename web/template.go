package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

func NewTemplate(layout Layout) *Template {
	return &Template{
		Funcs:  template.FuncMap{},
		Layout: layout,
	}
}

type Template struct {
	Layout       Layout
	DefaultTitle string
	Funcs        template.FuncMap
}

var logger = log.New(os.Stderr, "", 0)

func (t *Template) Handler(action Action) func(w http.ResponseWriter, r *http.Request) {
	if t.Layout == nil {
		return t.ActionHandler(action)
	} else {
		return t.LayoutHandler(t.Layout, action)
	}
}

func (t *Template) ActionHandler(action Action) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		action, ok := clone(action).(Action)
		if !ok {
			t.Render500(w, fmt.Errorf("type of action must be web.Action, was %T", action))
			return
		} else {
			t.HandleAction(w, r, action)
		}
	}
}

func renderAction(r *http.Request, action Action, funcs template.FuncMap) ([]byte, error) {
	b, e := action.Template()
	if e != nil {
		return nil, e
	}
	e = action.Load(r)
	if e != nil {
		return nil, e
	}
	return render(b, funcs, action)
}

func (t *Template) HandleAction(w http.ResponseWriter, r *http.Request, action Action) {
	b, e := renderAction(r, action, t.Funcs)
	if e != nil {
		t.Render500(w, e)
	} else {
		w.Write(b)
	}
}

// allow registering error pages
func (t *Template) Render500(w http.ResponseWriter, e error) {
	http.Error(w, e.Error(), 500)
}
