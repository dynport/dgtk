package confirm

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/dynport/dgtk/open"
)

func WebDiff(actions []*Action) (chan bool, error) {
	c := make(chan bool)
	mux := http.NewServeMux()
	mux.HandleFunc("/", webDiffHandler(actions))
	mux.HandleFunc("/_close", closeHandler(c))
	s := httptest.NewServer(mux)
	return c, open.URL(s.URL)
}

var actionColors = map[Type]string{
	TypeCreate: "lightgreen",
	TypeUpdate: "orange",
	TypeDelete: "pink",
}

func webDiffHandler(actions []*Action) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		buf := &bytes.Buffer{}
		ctx := &ActionsDiffCtx{Actions: actions}
		if err := ActionsDiff(buf, ctx); err != nil {
			http.Error(w, err.Error(), 500)
		} else {
			io.Copy(w, buf)
		}
	}
}

func closeHandler(c chan bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		switch s := strings.ToLower(r.Form.Get("submit")); s {
		case "confirm":
			if strings.ToLower(r.Form.Get("confirm")) == "confirm" {
				io.WriteString(w, "you can close this window now.")
				c <- true
			} else {
				http.Redirect(w, r, "/", http.StatusFound)
			}
			return
		default:
			io.WriteString(w, "you can close this window now.")
			c <- false
		}
	}
}

type ActionsDiffCtx struct {
	Actions []*Action
}

func (a *ActionsDiffCtx) ColorizeDiff(in []byte) string {
	lines := strings.Split(string(bytes.TrimSpace(in)), "\n")
	buf := &bytes.Buffer{}
	for _, l := range lines {
		if strings.HasPrefix(l, "-") {
			buf.WriteString(removeDiv(l))
		} else if strings.HasPrefix(l, "+") {
			buf.WriteString(addDiv(l))
		} else {
			buf.WriteString(div(l))
		}
	}
	return buf.String()
}

func (ctx *ActionsDiffCtx) ActionString(a *Action) string {
	txt := a.String()
	if a.Type == TypeDelete {
		return removeDiv(txt)
	} else if a.Type == TypeCreate {
		return addDiv(txt)
	}
	return div(txt)
}

func colorDiv(txt, color string) string {
	return `<div style="background-color:` + color + `">` + txt + `</div>`
}

func addDiv(txt string) string {
	return colorDiv(txt, "lightgreen")
}

func removeDiv(txt string) string {
	return colorDiv(txt, "pink")
}

func div(txt string) string {
	return `<div>` + txt + `</div>`
}
