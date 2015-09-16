package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/dynport/dgtk/dp-es/Godeps/_workspace/src/github.com/dynport/gocli"
	"github.com/dynport/dgtk/es"
)

type esIndexes struct {
	Host    string `cli:"opt -H default=127.0.0.1"`
	Compact bool   `cli:"opt --compact"`
}

func (r *esIndexes) Run() error {
	idx := &es.Index{Host: r.Host}
	parts := strings.Split(r.Host, ":")
	idx.Host = parts[0]
	var err error
	if len(parts) > 1 {
		idx.Port, err = strconv.Atoi(parts[1])
		if err != nil {
			return err
		}
	}
	stats, e := idx.Stats()
	if e != nil {
		return e
	}
	names := stats.IndexNames()
	if r.Compact {
		for _, n := range names {
			fmt.Println(n)
		}
		return nil
	}
	t := gocli.NewTable()
	if len(names) < 1 {
		logger.Printf("no indexes found")
		return nil
	}
	t.Add("name", "docs", "size")
	sort.Strings(names)
	for _, name := range names {
		index := stats.Indices[name]
		t.Add(name, index.Total.Docs.Count, sizePretty(index.Total.Store.SizeInBytes))
	}
	fmt.Println(t)
	return nil
}

func sizePretty(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.02fk", float64(size)/(1024.0))
	} else {
		return fmt.Sprintf("%.02fm", float64(size)/(1024.0*1024.0))
	}
}
