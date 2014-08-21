package web

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"reflect"
)

func (t *Template) LayoutHandler(layout Layout, action Action) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		e := func() error {
			action, ok := clone(action).(Action)
			if !ok {
				return fmt.Errorf("unable to cast %T into Action", action)
			}
			layout, ok := clone(layout).(Layout)
			if !ok {
				return fmt.Errorf("unable to cast %T into Layout", layout)
			}

			b, e := renderAction(r, action, t.Funcs)
			if e != nil {
				return e
			}

			e = layout.Load(r, action, template.HTML(b))
			if e != nil {
				return e
			}
			b, e = layout.Template()
			if e != nil {
				return e
			}
			b, e = render(b, t.Funcs, layout)
			if e != nil {
				return e
			}
			_, e = w.Write(b)
			if e != nil {
				logger.Printf("ERROR: %s", e)
			}
			return nil
		}()
		if e != nil {
			t.Render500(w, e)
		}
	}
}

func clone(p interface{}) interface{} {
	value := reflect.ValueOf(p)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	return reflect.New(value.Type()).Interface()
}

func render(b []byte, funcs template.FuncMap, i interface{}) ([]byte, error) {
	tpl, e := template.New(string(b)).Funcs(funcs).Parse(string(b))
	if e != nil {
		return nil, e
	}
	buf := &bytes.Buffer{}
	e = tpl.Execute(buf, i)
	return buf.Bytes(), e
}
