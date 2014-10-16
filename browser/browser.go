package browser

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

const UserAgentChrome = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.2062.120 Safari/537.36"

func New() (*Browser, error) {
	client := &http.Client{}
	t := &transport{client: client}
	client.Transport = t
	var e error
	client.Jar, e = cookiejar.New(nil)
	if e != nil {
		return nil, e
	}
	return &Browser{Client: client, transport: t}, nil
}

type transport struct {
	client    *http.Client
	userAgent string
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.userAgent != "" {
		r.Header.Set("User-Agent", t.userAgent)
	}
	r.Header.Set("Accept", "*/*")
	r.Header.Set("Host", r.URL.Host)
	return http.DefaultClient.Do(r)
}

type Logger interface {
	Printf(string, ...interface{})
}

type Browser struct {
	Logger
	*http.Client
	*http.Response
	transport *transport

	cachedBody []byte
}

func (b *Browser) UserAgent(ua string) {
	b.transport.userAgent = ua
}

func (b *Browser) Body() ([]byte, error) {
	if b.cachedBody == nil {
		if b.Response == nil {
			return nil, fmt.Errorf("Response was nil")
		}
		var e error
		b.cachedBody, e = ioutil.ReadAll(b.Response.Body)
		if e != nil {
			return nil, e
		}
	}
	return b.cachedBody, nil
}

func (b *Browser) Printf(format string, v ...interface{}) {
	if b.Logger != nil {
		b.Logger.Printf(format, v...)
	}
}

func (b *Browser) SaveToFile(path string) error {
	b.Printf("writing to file %q", path)
	body, e := b.Body()
	if e != nil {
		return e
	}
	return ioutil.WriteFile(path, body, 0644)
}

func (b *Browser) Submit(form *Form) error {
	values := url.Values{}
	for _, i := range form.Inputs {
		values.Set(i.Name, i.Value)
	}
	method := strings.ToUpper(form.Method)
	b.Printf("submitting form to %q with method %q", form.Action, method)
	f := values.Encode()
	req, e := http.NewRequest(method, form.Action, strings.NewReader(f))
	if e != nil {
		return e
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return b.Do(req)
}

func (b *Browser) Forms() ([]*Form, error) {
	body, e := b.Body()
	if e != nil {
		return nil, e
	}
	if b.Response == nil {
		return nil, fmt.Errorf("Response must not be nil")
	}
	return loadForms(b.Response.Request.URL.String(), body)
}

func (b *Browser) Visit(url string) error {
	req, e := http.NewRequest("GET", url, nil)
	if e != nil {
		return e
	}
	return b.Do(req)
}

func (b *Browser) Do(req *http.Request) error {
	if b.Response != nil {
		b.Response.Body.Close()
	}
	rsp, e := b.Client.Do(req)
	if e != nil {
		return e
	}
	if rsp.Status[0] != '2' {
		rsp.Body.Close()
		return fmt.Errorf("expected status 2xx, got %s", rsp.Status)
	}
	if b.Response != nil {
		b.Response.Body.Close()
	}
	b.cachedBody = nil
	b.Response = rsp
	return nil
}

func (b *Browser) Close() error {
	if b.Response != nil {
		return b.Response.Body.Close()
	}
	return nil
}
