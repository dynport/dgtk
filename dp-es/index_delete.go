package main

import "github.com/dynport/dgtk/es"

type indexDelete struct {
	Name string `cli:"arg required"`
	Host string `cli:"opt -H default=127.0.0.1"`
}

func (r *indexDelete) Run() error {
	idx := &es.Index{Host: r.Host, Index: r.Name}
	logger.Printf("deleting index %q", r.Name)
	return idx.DeleteIndex()
}
