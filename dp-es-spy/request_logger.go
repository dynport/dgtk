package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func requestLogger(h http.Handler) http.HandlerFunc {
	l := log.New(os.Stderr, "", 0)
	return func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()
		l.Printf("method=%s url=%s", r.Method, url)

		// only print body of search requests for now
		if strings.Contains(url, "/_search") {
			newBody := &bytes.Buffer{}
			capturedBody := &bytes.Buffer{}
			i, err := io.Copy(newBody, io.TeeReader(r.Body, capturedBody))
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			} else if i > 0 {
				io.WriteString(os.Stdout, capturedBody.String()+"\n")
			}
			r.Body = ioutil.NopCloser(newBody)
		}
		wr := &rw{ResponseWriter: w}
		started := time.Now()
		h.ServeHTTP(wr, r)
		l.Printf("status=%d total_time=%.06f", wr.status, time.Since(started).Seconds())
		if wr.status >= 400 {
			io.Copy(os.Stdout, wr.buf)
			io.WriteString(os.Stdout, "\n")
		}
		io.WriteString(os.Stdout, strings.Repeat("-", 100)+"\n")
	}
}
