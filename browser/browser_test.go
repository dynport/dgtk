package browser

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestName(t *testing.T) {
	Convey("Cookie Handling and User-Agent", t, func() {
		browser, e := New()
		So(e, ShouldBeNil)
		So(browser, ShouldNotBeNil)

		s := httptest.NewServer(http.HandlerFunc(testHandler))

		u := "http://" + s.Listener.Addr().String()

		e = browser.Visit(u)
		So(e, ShouldBeNil)

		b, e := browser.Body()
		So(e, ShouldBeNil)
		So(string(b), ShouldContainSubstring, `ua="Go 1.1 package http"`)

		// validate that the correct user agent
		// and the correct cookies are set
		browser.UserAgent("GoogleBot")
		e = browser.Visit(u)
		logger.Printf("reading body")
		b, e = browser.Body()
		So(e, ShouldBeNil)
		So(string(b), ShouldContainSubstring, "cookie=value")
		So(string(b), ShouldContainSubstring, "GoogleBot")
	})
}

var logger = log.New(os.Stderr, "", 0)

func testHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Path: "/", Name: "cookie", Value: "value"})
	out := []string{fmt.Sprintf("ua=%q", r.Header.Get("User-Agent"))}
	if c, e := r.Cookie("cookie"); e == nil {
		out = append(out, fmt.Sprintf("cookie=%s", c.Value))
	}
	fmt.Fprintf(w, strings.Join(out, " "))
}
