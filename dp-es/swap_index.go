package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type swapIndex struct {
	Host      string `cli:"opt -H default=127.0.0.1"`
	NewIndex  string `cli:"arg required"`
	AliasName string `cli:"arg required"`
}

func (r *swapIndex) Run() error {
	all, e := indexAliases(r.Host)
	if e != nil {
		return e
	}

	actions := list{}

	for idx, alias := range all {
		if _, ok := alias.Aliases[r.AliasName]; ok {
			actions = append(actions, hash{"remove": hash{"index": idx, "alias": r.AliasName}})
		}
	}

	actions = append(actions, hash{"add": hash{"index": r.NewIndex, "alias": r.AliasName}})

	b, e := json.Marshal(hash{"actions": actions})
	if e != nil {
		return e
	}
	req, e := http.NewRequest("POST", "http://"+r.Host+":9200/_aliases", bytes.NewReader(b))
	if e != nil {
		return e
	}
	rsp, e := http.DefaultClient.Do(req)
	if e != nil {
		return e
	}
	b, e = ioutil.ReadAll(rsp.Body)
	if rsp.Status[0] != '2' {
		return fmt.Errorf("expected status 2xx, got %s: %s", rsp.Status, string(b))
	}
	return nil
}
