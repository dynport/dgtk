package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type dump struct {
	IndexName      string `cli:"arg required"`
	Address        string `cli:"opt -a default=http://127.0.0.1:9200"`
	BatchSize      int    `cli:"opt -b default=1000"`
	ScrollDuration string `cli:"opt -s default=1m"`
}

func (r *dump) Run() error {
	docs, err := iterateIndex(r.Address, r.IndexName, r.BatchSize, r.ScrollDuration)
	if err != nil {
		return err
	}

	c, wg := progress()
	for d := range docs {
		if _, err = io.WriteString(os.Stdout, string(d)+"\n"); err != nil {
			return err
		}
		c <- struct{}{}
	}
	close(c)
	wg.Wait()
	return nil
}

func progress() (chan struct{}, *sync.WaitGroup) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	c := make(chan struct{})
	go func() {
		defer wg.Done()
		t := time.Tick(1 * time.Second)
		started := time.Now()
		cnt := 0

		printStatus := func() {
			diff := time.Since(started).Seconds()
			perSecond := float64(cnt) / diff
			logger.Printf("cnt=%d time=%.01f per_second=%.1f/second", cnt, diff, perSecond)
		}
		for {
			select {
			case _, ok := <-c:
				if !ok {
					printStatus()
					return
				}
				cnt++
			case <-t:
				printStatus()
			}

		}
	}()
	return c, wg
}

func iterateIndex(addr, name string, size int, scroll string) (chan json.RawMessage, error) {
	scrollID, err := openIndex(addr, name, size, scroll)
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

func openIndex(addr, name string, size int, scroll string) (scrollID string, err error) {
	rsp, err := http.Post(addr+"/"+name+"/_search?search_type=scan&scroll="+scroll,
		"application/json",
		strings.NewReader(`{"query": {"match_all": {} }, "size": `+strconv.Itoa(size)+`}`),
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
