package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type dump struct {
	IndexName      string   `cli:"arg required"`
	Address        string   `cli:"opt -a default=http://127.0.0.1:9200"`
	BatchSize      int      `cli:"opt -b default=1000"`
	ScrollDuration string   `cli:"opt -s default=1m"`
	Fields         []string `cli:"opt --fields"`
}

func (r *dump) Run() error {
	docs, err := iterateIndex(r.Address, r.IndexName, OpenIndexSize(r.BatchSize), OpenIndexScroll(r.ScrollDuration), OpenIndexFields(r.Fields))
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

type openIndexOpt struct {
	Size   int
	Scroll string
	Fields []string
}

func OpenIndexSize(size int) func(*openIndexOpt) {
	return func(o *openIndexOpt) {
		o.Size = size
	}
}

func OpenIndexScroll(scroll string) func(*openIndexOpt) {
	return func(o *openIndexOpt) {
		o.Scroll = scroll
	}
}

func OpenIndexFields(fields []string) func(*openIndexOpt) {
	return func(o *openIndexOpt) {
		o.Fields = fields
	}
}

func iterateIndex(addr, name string, funcs ...func(*openIndexOpt)) (chan json.RawMessage, error) {
	scrollID, err := openIndex(addr, name, funcs...)
	if err != nil {
		return nil, err
	}
	c := make(chan json.RawMessage)
	var docs []json.RawMessage
	go func() {
		defer close(c)
		for {
			scrollID, docs, err = loadDocumentsWithScroll(addr, scrollID)
			if err != nil {
				logger.Printf("err=%q", err)
				return
			} else if len(docs) == 0 {
				return
			}
			for _, d := range docs {
				c <- d
			}
		}
	}()
	return c, nil
}

type scrollResponse struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Hits []json.RawMessage `json:"hits"`
	} `json:"hits"`
}

func loadDocumentsWithScroll(addr string, id string) (string, []json.RawMessage, error) {
	req, err := http.NewRequest("GET", addr+"/_search/scroll?scroll=1m", strings.NewReader(id))
	if err != nil {
		return "", nil, err
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer rsp.Body.Close()
	switch rsp.StatusCode {
	case 200:
		var rr *scrollResponse
		err = json.NewDecoder(rsp.Body).Decode(&rr)
		if err != nil {
			return "", nil, err
		}
		return rr.ScrollID, rr.Hits.Hits, nil
	default:
		b, _ := ioutil.ReadAll(rsp.Body)
		return "", nil, fmt.Errorf("got status %s but expected 2x. body=%s", rsp.Status, string(b))
	}
}

type openIndexDoc struct {
	Size   int      `json:"size"`
	Fields []string `json:"fields,omitempty"`
}

func openIndex(addr, name string, funcs ...func(*openIndexOpt)) (scrollID string, err error) {
	o := &openIndexOpt{}
	for _, f := range funcs {
		f(o)
	}
	doc := &openIndexDoc{
		Size:   o.Size,
		Fields: o.Fields,
	}
	b, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}

	rsp, err := http.Post(addr+"/"+name+"/_search?search_type=scan&scroll="+o.Scroll,
		"application/json",
		bytes.NewReader(b),
	)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		b, _ := ioutil.ReadAll(rsp.Body)
		return "", fmt.Errorf("expected status 2xx, got %s: %s", rsp.Status, string(b))
	}
	var s *scrollResponse
	err = json.NewDecoder(rsp.Body).Decode(&s)
	if err != nil {
		return "", err
	}
	return s.ScrollID, nil
}
