package wunderproxy

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWunderproxyBase(t *testing.T) {
	proxyHandler := NewProxy()
	proxy := httptest.NewServer(proxyHandler)

	s1 := httptest.NewServer(&testServer{id: "server 1"})
	defer s1.Close()
	s1Address := strings.TrimPrefix(s1.URL, "http://")
	s2 := httptest.NewServer(&testServer{id: "server 2"})
	defer s2.Close()
	s2Address := strings.TrimPrefix(s2.URL, "http://")

	resp, e := http.Get(proxy.URL)
	if e != nil {
		t.Fatalf("failed to send request to proxy: %s", e)
	}
	resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404 status code, as no address set yet")
	}

	proxyHandler.Update(s1Address, "")
	testRequest(t, proxy, "GET", "GET server 1")
	testRequest(t, proxy, "PUT", "PUT server 1")
	testRequest(t, proxy, "POST", "POST server 1")
	testRequest(t, proxy, "DELETE", "DELETE server 1")

	proxyHandler.Update(s2Address, "")
	testRequest(t, proxy, "GET", "GET server 2")
	testRequest(t, proxy, "PUT", "PUT server 2")
	testRequest(t, proxy, "POST", "POST server 2")
	testRequest(t, proxy, "DELETE", "DELETE server 2")

	proxyHandler.Update(s1Address, "")
	testRequest(t, proxy, "GET", "GET server 1")

	t.Logf("yeah did it")
}

func testRequest(t *testing.T, proxy *httptest.Server, method, expectation string) {
	req, e := http.NewRequest(method, proxy.URL, nil)
	if e != nil {
		t.Fatalf("failed to create request to proxy: %s", e)
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		t.Fatalf("failed to send request to proxy: %s", e)
	}
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		t.Fatalf("failed to read response from proxy with server 1 set")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 333 {
		t.Errorf("expected 333 status code, as address set. Got: %d", resp.StatusCode)
	}
	result := strings.TrimSpace(string(body))
	if result != expectation {
		t.Errorf("expected %s\ngot:      %q", expectation, result)
	}
}

type testServer struct {
	id string
}

func (ts *testServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, r.Method+" "+ts.id, 333)
}

func TestSaneTotalTimePrinter(t *testing.T) {
	tt := []struct {
		secs float64
		exp  string
	}{
		{1, "1.000000s"},
		{0.1, "0.100000s"},
		{0.01, "0.010000s"},
		{0.001, "0.001000s"},
		{0.0001, "0.100000ms"},
		{0.0002, "0.200000ms"},
		{0.00099, "0.990000ms"},
		{0.00001, "0.010000ms"},
	}

	for i := range tt {
		got := saneTotalTimePrinter(tt[i].secs)
		if got != tt[i].exp {
			t.Errorf("in test %d: expected %q, got %q", i, tt[i].exp, got)
		}
	}
}
