package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "index")
}

func namedRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "named route")
}

func paramsAction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Write([]byte(""))
}

type userShow struct {
	Id string
}

func (t *userShow) Load(r *http.Request, params httprouter.Params) error {
	t.Id = params.ByName("id")
	return nil
}

func (t *userShow) Template() ([]byte, error) {
	return []byte("User {{ .Id }} => named_route={{ namedPath }}"), nil
}

func TestRouter(t *testing.T) {
	var ex, v interface{}
	app := &App{}
	r := &Router{
		"GET": {
			"/":                  index,
			"/params/:id/action": paramsAction,
			"/user/:id/show":     &userShow{},
			"/index":             &NamedRoute{Name: "named", Handler: namedRoute},
		},
	}

	r2, e := r.Router(app)
	if e != nil {
		t.Fatal("error getting router", e)
	}

	for _, s := range []string{"userShowPath", "namedPath"} {
		if _, ok := app.Funcs[s]; !ok {
			t.Errorf("func %v should exist", s)
		}
	}
	f, ok := app.Funcs["userShowPath"].(UrlFunc)
	if !ok {
		t.Errorf("func userShowPath should exist")
	}
	ex = "/user/123/show"
	v = f(123)

	if v != ex {
		t.Errorf("f(123) should eq %v, was %v", ex, v)
	}

	ex = "/user/124/show"
	v = app.Url("userShow", 124)
	if v != ex {
		t.Errorf(`app.Url("userShow", 124) should eq %v, was %v`, ex, v)
	}
	s := httptest.NewServer(r2)
	defer s.Close()
	addr := s.Listener.Addr()

	var i interface{} = &userShow{}

	_, ok = i.(Action)
	if !ok {
		t.Errorf("userShow should cast to Action")
	}

	rsp, body, e := get("http://" + addr.String())
	if rsp.StatusCode != 200 {
		t.Errorf("expected StatusCode to eq 200, was %v", rsp.StatusCode)
	}

	if body != "index" {
		t.Errorf("expected body to eq index, was %v", body)
	}

	rsp, body, e = get("http://" + addr.String() + "/index")

	if rsp.StatusCode != 200 {
		t.Errorf("expected StatusCode to eq 200, was %v", rsp.StatusCode)
	}
	if body != "named route" {
		t.Errorf("expected body to eq %q, was %q", "named route", body)
	}

	rsp, body, e = get("http://" + addr.String() + "/user/12/show")
	if rsp.StatusCode != 200 {
		t.Errorf("expected StatusCode to eq 200, was %v", rsp.StatusCode)
	}
	ex = "User 12 => named_route=/index"
	if ex != body {
		t.Errorf("expected body to eq %v, was %v", ex, body)
	}
}

func get(url string) (*http.Response, string, error) {
	rsp, e := http.Get(url)
	if e != nil {
		return nil, "", e
	}
	defer rsp.Body.Close()
	b, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return nil, "", e
	}
	return rsp, string(b), nil
}
