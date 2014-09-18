package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	. "github.com/smartystreets/goconvey/convey"
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
	return []byte("User {{ .Id }} => named_route={{ named }}"), nil
}

func TestRouter(t *testing.T) {
	Convey("Router", t, func() {
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
		So(e, ShouldBeNil)
		So(r2, ShouldNotBeNil)

		So(app.Funcs["userShow"], ShouldNotBeNil)
		So(app.Funcs["named"], ShouldNotBeNil)
		f, ok := app.Funcs["userShow"].(UrlFunc)
		So(f, ShouldNotBeNil)
		So(ok, ShouldEqual, true)
		So(f(123), ShouldEqual, "/user/123/show")
		So(app.Url("userShow", 124), ShouldEqual, "/user/124/show")

		s := httptest.NewServer(r2)
		defer s.Close()
		addr := s.Listener.Addr()

		var i interface{} = &userShow{}

		_, ok = i.(Action)
		So(ok, ShouldEqual, true)

		rsp, body, e := get("http://" + addr.String())
		So(rsp.StatusCode, ShouldEqual, 200)
		So(body, ShouldEqual, "index")

		rsp, body, e = get("http://" + addr.String() + "/index")
		So(rsp.StatusCode, ShouldEqual, 200)
		So(body, ShouldEqual, "named route")

		rsp, body, e = get("http://" + addr.String() + "/user/12/show")
		So(rsp.StatusCode, ShouldEqual, 200)
		So(body, ShouldEqual, "User 12 => named_route=/index")
	})
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
