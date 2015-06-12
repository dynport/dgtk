package main

import (
	"io"
	"os"

	"github.com/dynport/dgtk/es"
)

type dump struct {
	IndexName      string   `cli:"arg required"`
	Address        string   `cli:"opt -a default=http://127.0.0.1:9200"`
	BatchSize      int      `cli:"opt -b default=1000"`
	ScrollDuration string   `cli:"opt -s default=1m"`
	Fields         []string `cli:"opt --fields"`
}

func (r *dump) Run() error {
	docs, err := es.IterateIndex(r.Address, r.IndexName, es.OpenIndexSize(r.BatchSize), es.OpenIndexScroll(r.ScrollDuration), es.OpenIndexFields(r.Fields))
	if err != nil {
		return err
	}

	for d := range docs {
		if _, err = io.WriteString(os.Stdout, string(d)+"\n"); err != nil {
			return err
		}
	}
	return nil
}
