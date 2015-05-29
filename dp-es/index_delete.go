package main

import "github.com/dynport/dgtk/es"

type indexDelete struct {
	Names []string `cli:"arg required"`
	Host  string   `cli:"opt -H default=127.0.0.1"`
}

func (r *indexDelete) Run() error {
	for _, n := range r.Names {
		idx := &es.Index{Host: r.Host, Index: n}
		logger.Printf("deleting index %q", n)
		if err := idx.DeleteIndex(); err != nil {
			return err
		}
	}
	return nil
}
