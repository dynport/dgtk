package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dynport/dgtk/dp-es/Godeps/_workspace/src/github.com/dynport/gocli"
)

type esAliases struct {
	Host string `cli:"opt -H default=127.0.0.1"`
}

func indexAliases(host string) (map[string]*IndexAlias, error) {
	rsp, e := http.Get("http://" + host + ":9200/_aliases")
	if e != nil {
		return nil, e
	}
	b, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return nil, e
	}
	var m map[string]*IndexAlias
	e = json.Unmarshal(b, &m)
	if e != nil {
		return nil, e
	}
	return m, nil
}

func (r *esAliases) Run() error {
	m, e := indexAliases(r.Host)
	if e != nil {
		return e
	}
	t := gocli.NewTable()
	for name, a := range m {
		aliases := []string{}
		for k := range a.Aliases {
			aliases = append(aliases, k)
		}
		t.Add(name, strings.Join(aliases, ", "))
	}
	fmt.Println(t)
	return nil
}

type IndexAlias struct {
	Aliases map[string]interface{} `json:"aliases"`
}
