package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type aliasCreate struct {
	Host      string `cli:"opt -H default=http://127.0.0.1:9200"`
	Index     string `cli:"arg required"`
	AliasName string `cli:"arg required"`
}

type hash map[string]interface{}

type list []interface{}

func (r *aliasCreate) Run() error {
	h := hash{
		"actions": list{
			hash{"add": hash{"index": r.Index, "alias": r.AliasName}},
		},
	}
	b, e := json.Marshal(h)
	if e != nil {
		return e
	}
	req, e := http.NewRequest("POST", normalizeIndexAddress(r.Host)+"/_aliases", bytes.NewReader(b))
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
