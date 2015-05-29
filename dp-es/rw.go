package main

import (
	"bytes"
	"io"
	"net/http"
)

type rw struct {
	http.ResponseWriter
	status int
	buf    *bytes.Buffer
}

func (w *rw) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *rw) Write(b []byte) (int, error) {
	if w.buf == nil {
		w.buf = &bytes.Buffer{}
	}
	return io.MultiWriter(w.ResponseWriter, w.buf).Write(b)
}
