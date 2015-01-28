package awsv4

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"
)

const (
	iso8601BasicFormat      = "20060102T150405Z"
	iso8601BasicFormatShort = "20060102"
	lf                      = "\n"
)

// Context encapsulates the context of a client's connection to an AWS service.
type Context struct {
	Service string
	Region  string

	AccessKeyID     string
	SecretAccessKey string
	SecurityToken   string

	CurrentTime time.Time
}

func (c *Context) currentTime() time.Time {
	if c.CurrentTime.IsZero() {
		return time.Now()
	}
	return c.CurrentTime
}

func (c *Context) Sign(r *http.Request) error {
	date := r.Header.Get("Date")
	t := c.currentTime().UTC()
	if date != "" {
		var err error
		t, err = time.Parse(http.TimeFormat, date)
		if err != nil {
			return err
		}
	}
	r.Header.Set("x-amz-date", t.Format(iso8601BasicFormat))

	chash, err := payloadHash(r)
	if err != nil {
		return err
	}
	r.Header.Set("x-amz-content-sha256", chash)

	if s := c.SecurityToken; s != "" {
		r.Header.Set("X-Amz-Security-Token", s)
	}

	k := c.signature(c.SecretAccessKey, t)
	h := hmac.New(sha256.New, k)
	c.writeStringToSign(h, t, r, chash)

	auth := bytes.NewBufferString("AWS4-HMAC-SHA256 ")
	io.WriteString(auth,
		strings.Join(
			[]string{
				"Credential=" + c.AccessKeyID + "/" + c.creds(t),
				",",
				" ",
				"SignedHeaders=",
			},
			"",
		),
	)

	c.writeHeaderList(auth, r)
	io.WriteString(auth, ", Signature="+fmt.Sprintf("%x", h.Sum(nil)))
	r.Header.Set("Authorization", auth.String())
	return nil
}

func (c *Context) writeStringToSign(w io.Writer, t time.Time, r *http.Request, chash string) {
	io.WriteString(w,
		strings.Join(
			[]string{"AWS4-HMAC-SHA256", t.Format(iso8601BasicFormat), c.creds(t), ""},
			lf,
		),
	)
	h := sha256.New()
	c.writeRequest(h, r, chash)
	fmt.Fprintf(w, "%x", h.Sum(nil))
}

func (c *Context) writeRequest(w io.Writer, r *http.Request, chash string) {
	r.Header.Set("host", r.Host)

	io.WriteString(w, r.Method+lf)
	c.writeURI(w, r)
	io.WriteString(w, lf)
	c.writeQuery(w, r)
	io.WriteString(w, lf)
	c.writeHeader(w, r)
	io.WriteString(w, lf+lf)
	c.writeHeaderList(w, r)
	io.WriteString(w, lf+chash)
}

func (c *Context) writeURI(w io.Writer, r *http.Request) {
	p := r.URL.RequestURI()
	if r.URL.RawQuery != "" {
		p = p[:len(p)-len(r.URL.RawQuery)-1]
	}
	slash := strings.HasSuffix(p, "/")
	p = path.Clean(p)
	if p != "/" && slash {
		p += "/"
	}
	io.WriteString(w, p)
}

func (c *Context) writeQuery(w io.Writer, r *http.Request) {
	var a []string
	for k, vs := range r.URL.Query() {
		k = url.QueryEscape(k)
		for _, v := range vs {
			if v == "" {
				a = append(a, k+"=")
			} else {
				v = url.QueryEscape(v)
				a = append(a, k+"="+v)
			}
		}
	}
	sort.Strings(a)
	io.WriteString(w, strings.Join(a, "&"))
}

func (c *Context) writeHeader(w io.Writer, r *http.Request) {
	i, a := 0, make([]string, len(r.Header))
	for k, v := range r.Header {
		sort.Strings(v)
		a[i] = strings.ToLower(k) + ":" + strings.Join(v, ",")
		i++
	}
	sort.Strings(a)
	io.WriteString(w, strings.Join(a, lf))
}

func (c *Context) writeHeaderList(w io.Writer, r *http.Request) {
	i, a := 0, make([]string, len(r.Header))
	for k := range r.Header {
		a[i] = strings.ToLower(k)
		i++
	}
	sort.Strings(a)
	io.WriteString(w, strings.Join(a, ";"))
}

func (c *Context) creds(t time.Time) string {
	return t.Format(iso8601BasicFormatShort) + "/" + c.Region + "/" + c.Service + "/aws4_request"
}

// TODO: figure out if this is actually needed, maybe try the first n bytes and then ignore the checksum?!?
func payloadHash(r *http.Request) (string, error) {
	// If the payload is empty, use the empty string as the input to the SHA256 function
	// http://docs.amazonwebservices.com/general/latest/gr/sigv4-create-canonical-request.html
	h := sha256.New()
	var body io.Reader = r.Body
	if r.Body == nil {
		body = ioutil.NopCloser(strings.NewReader(""))
	}
	buf := &bytes.Buffer{}
	reader := io.TeeReader(body, buf)
	io.Copy(h, reader)
	if r.Body != nil {
		r.Body = ioutil.NopCloser(buf)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (c *Context) signature(secretAccessKey string, t time.Time) []byte {
	h := ghmac(
		[]byte("AWS4"+secretAccessKey),
		t.Format(iso8601BasicFormatShort),
	)
	h = ghmac(h, c.Region)
	h = ghmac(h, c.Service)
	h = ghmac(h, "aws4_request")
	return h
}

func ghmac(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	io.WriteString(h, data)
	return h.Sum(nil)
}
