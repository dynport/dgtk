package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type spy struct {
	ESAddress string `cli:"arg required desc='Address of Elasticsearch host to connect'"`
	Address   string `cli:"opt -a default=127.0.0.1:9201 desc='Address to bind to'"`
}

func (r *spy) Run() error {
	l := log.New(os.Stderr, "", 0)
	u, err := url.Parse(r.ESAddress)
	if err != nil {
		return err
	}
	l.Printf("starting on addr %q, proxying to %q", r.Address, r.ESAddress)
	return http.ListenAndServe(r.Address, requestLogger(httputil.NewSingleHostReverseProxy(u)))
}
