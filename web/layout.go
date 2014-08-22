package web

import (
	"fmt"
	"html/template"
	"net/http"
)

type Layout interface {
	Load(*http.Request, Action, template.HTML) error
	Template() ([]byte, error)
}

func (t *App) LayoutHandler(layout Layout, action Action) func(w http.ResponseWriter, r *http.Request) {
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
			t.HandleError(w, e)
		}
	}
}
