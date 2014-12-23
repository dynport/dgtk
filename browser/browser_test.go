package browser

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestAbsoluteUrl(t *testing.T) {
	u, err := url.Parse("http://www.mysite.com/some/path")
	if err != nil {
		t.Fatalf("error parsing initial url: %s", err)
	}

	tests := []struct {
		Path   string
		Result string
	}{
		{"some.action.html", "http://www.mysite.com/some/path/some.action.html"},
		{"/index.html", "http://www.mysite.com/index.html"},
		{"http://some.other.site.de", "http://some.other.site.de"},
	}
	for _, tst := range tests {
		abs := AbsoluteURL(tst.Path, u)
		if abs != tst.Result {
			t.Errorf("expected absolute url for path %q to be %q, was %q", tst.Path, tst.Result, abs)
		}
	}
}

func TestName(t *testing.T) {
	browser, err := New()
	if err != nil {
		t.Fatal("error initializing browser: %s", err)
	}
	s := httptest.NewServer(http.HandlerFunc(testHandler))

	u := "http://" + s.Listener.Addr().String()

	err = browser.Visit(u)
	if err != nil {
		t.Fatal("error visiting %q: %s", u, err)
	}

	b, err := browser.Body()
	if err != nil {
		t.Fatalf("error getting body of browser", err)
	}
	bs := string(b)
	if !strings.Contains(bs, `ua="Go 1.1 package http"`) {
		t.Errorf("expected body %q to contain %q", bs, `ua="Go 1.1 package http"`)
	}
	// validate that the correct user agent
	// and the correct cookies are set
	browser.UserAgent("GoogleBot")
	if err := browser.Visit(u); err != nil {
		t.Fatalf("error visiting url %s: %s", u, err)
	}
	b, err = browser.Body()
	if err != nil {
		t.Fatalf("error gettong bod", err)

	}
	str := string(b)
	for _, p := range []string{"cookie=value", "GoogleBot"} {
		if !strings.Contains(str, p) {
			t.Errorf("expected body %q to contain %q", str, p)
		}
	}
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
