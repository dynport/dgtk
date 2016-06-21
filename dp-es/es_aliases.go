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
	Host string `cli:"opt -H default=http://127.0.0.1:9200"`
}

func indexAliases(address string) (map[string]*IndexAlias, error) {
	rsp, e := http.Get(address + "/_aliases")
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
	m, e := indexAliases(normalizeIndexAddress(r.Host))
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
