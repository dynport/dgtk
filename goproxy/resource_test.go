package goproxy

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

type testHandler struct {
}

func (handler *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("hello world"))
}

const addr = "127.0.0.1:11223"

func testServer(t *testing.T) (net.Listener, error) {
	l, e := net.Listen("tcp", addr)
	if e != nil {
		t.Fatal(e.Error())
	}
	go func() {
		http.Serve(l, &testHandler{})
	}()
	timeout := time.NewTimer(1 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	for {
		select {
		case <-timeout.C:
			return nil, fmt.Errorf("timeout waiting for http server")
		case <-ticker.C:
			_, e := net.DialTimeout("tcp", addr, 1*time.Millisecond)
			if e == nil {
				return l, nil
			}
		}
	}
}

func newGetResource(t *testing.T, theUrl string) *Resource {
	u, e := url.Parse(theUrl)
	if e != nil {
		t.Fatal(e.Error())
	}
	return &Resource{Method: "GET", Url: u}
}

func TestResource(t *testing.T) {
	DefaultCacheDir = "./tmp"
	l, e := testServer(t)
	if e != nil {
		t.Fatal(e.Error())
	}
	defer l.Close()
	Convey("Resource", t, func() {
		os.RemoveAll("./tmp")
		os.MkdirAll("./tmp", 0755)
		proxy, e := New("./tmp")
		if e != nil {
			t.Fatal(e.Error())
		}

		u, e := url.Parse("http://" + addr)
		So(e, ShouldBeNil)
		r := &Resource{Method: "GET", Url: u}

		So(proxy.cached(r), ShouldBeFalse)
		fetched, e := proxy.Load(r)
		So(fetched, ShouldBeTrue)
		So(e, ShouldBeNil)

		So(proxy.Store(r), ShouldBeNil)
		So(proxy.cached(r), ShouldBeTrue)

		proxy.Load(r)
		fetched, e = proxy.Load(r)
		So(fetched, ShouldBeFalse)
		So(e, ShouldBeNil)
		So(proxy.cached(r), ShouldBeTrue)
		So(r.Response.FetchedAt.UnixNano(), ShouldBeGreaterThan, time.Now().Add(-1*time.Second).UnixNano())
		So(string(r.Response.Body), ShouldEqual, "hello world")
	})
}
