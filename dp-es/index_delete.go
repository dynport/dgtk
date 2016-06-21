package main

import "github.com/dynport/dgtk/es"

type indexDelete struct {
	Names []string `cli:"arg required"`
	Host  string   `cli:"opt -H default=http://127.0.0.1:9200"`
}

func (r *indexDelete) Run() error {
	addr := normalizeIndexAddress(r.Host)
	for _, n := range r.Names {
		idx := &es.Index{Address: addr}
		logger.Printf("deleting index %q", n)
		if err := idx.DeleteIndex(); err != nil {
			return err
		}
	}
	return nil
}
