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

func TestResource(t *testing.T) {
	DefaultCacheDir = "./tmp"
	l, e := testServer(t)
	if e != nil {
		t.Fatal(e.Error())
	}
	defer l.Close()
	Convey("Resource", t, func() {
		os.RemoveAll("./tmp")
		u, e := url.Parse("http://" + addr)
		So(e, ShouldBeNil)
		r := &Resource{Method: "GET", Url: u}

		Convey("cachePath", func() {
			So(r.cachePath(), ShouldEqual, "./tmp/127.0.0.1:11223/index")
		})

		So(r.checksum(), ShouldEqual, "1bbb8d4bf13fee2dd452cab0019ed1ee")
		So(r.cachePath(), ShouldEqual, "./tmp/127.0.0.1:11223/index")
		So(r.cached(), ShouldBeFalse)
		fetched, e := r.Load()
		So(fetched, ShouldBeTrue)
		So(e, ShouldBeNil)

		So(r.store(), ShouldBeNil)
		t.Log(r.cachePath())
		So(r.cached(), ShouldBeTrue)

		r.Load()
		fetched, e = r.Load()
		So(fetched, ShouldBeFalse)
		So(e, ShouldBeNil)
		So(r.cached(), ShouldBeTrue)
		So(r.Response.FetchedAt.UnixNano(), ShouldBeGreaterThan, time.Now().Add(-1*time.Second).UnixNano())
	})
}
