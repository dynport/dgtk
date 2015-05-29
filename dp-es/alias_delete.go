package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type aliasDelete struct {
	Host      string `cli:"opt -H default=127.0.0.1"`
	Index     string `cli:"arg required"`
	AliasName string `cli:"arg required"`
}

func (r *aliasDelete) Run() error {
	h := hash{
		"actions": list{
			hash{"remove": hash{"index": r.Index, "alias": r.AliasName}},
		},
	}
	b, e := json.Marshal(h)
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
