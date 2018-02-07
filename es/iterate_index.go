package es

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

func OpenIndexSize(size int) func(*openIndexOpt) {
	return func(o *openIndexOpt) {
		o.Size = size
	}
}

type openIndexOpt struct {
	Size   int
	Scroll string
	Fields []string
	Query  *Query
}

// timespan how long each request is valid (e.g. 1m)
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

func OpenIndexQuery(query *Query) func(*openIndexOpt) {
	return func(o *openIndexOpt) {
		o.Query = query
	}
}

func IterateIndex(addr, name string, funcs ...func(*openIndexOpt)) (chan json.RawMessage, error) {
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
				log.Printf("%+v", err)
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

type openIndexDoc struct {
	Size   int      `json:"size"`
	Fields []string `json:"fields,omitempty"`
	Query  *Query   `json:"query,omitempty"`
}

func openIndex(addr, name string, funcs ...func(*openIndexOpt)) (scrollID string, err error) {
	o := &openIndexOpt{Scroll: "1m", Size: 1000}
	for _, f := range funcs {
		f(o)
	}
	doc := &openIndexDoc{
		Size:   o.Size,
		Fields: o.Fields,
		Query:  o.Query,
	}
	b, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}

	rsp, err := http.Post(addr+"/"+name+"/_search?&scroll="+o.Scroll,
		"application/json",
		bytes.NewReader(b),
	)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		b, _ := ioutil.ReadAll(rsp.Body)
		return "", errors.Wrapf(err, "opening index: expected status 2xx, got %s: %s", rsp.Status, string(b))
	}
	var s *scrollResponse
	err = json.NewDecoder(rsp.Body).Decode(&s)
	if err != nil {
		return "", err
	}
	return s.ScrollID, nil
}

func loadDocumentsWithScroll(addr string, id string) (string, []json.RawMessage, error) {
	pl := struct {
		Scroll   string `json:"scroll"`
		ScrollID string `json:"scroll_id"`
	}{
		"1m", id,
	}
	b, err := json.Marshal(pl)
	if err != nil {
		return "", nil, err
	}
	rsp, err := http.Post(addr+"/_search/scroll", "application/json", bytes.NewReader(b))
	if err != nil {
		return "", nil, err
	}
	defer rsp.Body.Close()
	switch rsp.StatusCode {
	case 200:
		var rr *scrollResponse
		err = json.NewDecoder(io.TeeReader(rsp.Body, os.Stdout)).Decode(&rr)
		if err != nil {
			return "", nil, err
		}
		return rr.ScrollID, rr.Hits.Hits, nil
	default:
		b, _ := ioutil.ReadAll(rsp.Body)
		return "", nil, errors.Errorf("load documents with scroll: got status %s but expected 2x. body=%s", rsp.Status, string(b))
	}
}

type scrollResponse struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Hits []json.RawMessage `json:"hits"`
	} `json:"hits"`
}
