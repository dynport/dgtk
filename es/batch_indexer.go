package es

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func NewBatchIndexer(addr string) *BatchIndexer {
	i := &BatchIndexer{
		Address:       addr,
		buf:           &bytes.Buffer{},
		doClose:       make(chan struct{}),
		closed:        make(chan struct{}),
		docs:          make(chan *Doc),
		flushCount:    1000,
		flushDuration: 1 * time.Second,
	}
	go i.start()
	return i
}

type BatchIndexer struct {
	Address       string
	buf           *bytes.Buffer
	cnt           int
	docs          chan *Doc
	closed        chan struct{}
	doClose       chan struct{}
	lastFlush     time.Time
	flushDuration time.Duration
	flushCount    int
}

func (i *BatchIndexer) Add(d *Doc) error {
	if d.Id == "" {
		return errors.New("ID must be set")
	}
	if d.Index == "" {
		return errors.New("Index must be set")
	}
	if d.Type == "" {
		return errors.New("Type must be set")
	}
	i.docs <- d
	return nil
}

func (i *BatchIndexer) Close() error {
	close(i.doClose)
	<-i.closed
	return nil
}

func (i *BatchIndexer) start() {
	defer close(i.closed)
	t := time.NewTicker(1 * time.Second)
	for {
		select {
		case d := <-i.docs:
			// add doc to buffer
			_, err := i.addInternal(d)
			if err != nil {
				log.Printf("err=%q", err)
			}
			i.checkFlush()
		case <-t.C:
			// index
			i.checkFlush()
		case <-i.doClose:
			// should stop indexing
			i.flush()
			return
		}
	}
}

// checkFlush calls flush if
// a) time since last flush > threshold
// or
// b) rows since last flush > threshold
func (i *BatchIndexer) checkFlush() bool {
	if time.Since(i.lastFlush) < i.flushDuration && i.cnt < i.flushCount {
		return false
	}
	i.flush()
	return true
}

func (i *BatchIndexer) flush() {
	if i.cnt == 0 || i.buf.Len() == 0 {
		return
	}
	defer func() {
		i.cnt = 0
	}()
	b := i.buf
	i.buf = &bytes.Buffer{}
	err := func() error {
		u := i.Address + "/_bulk"
		req, err := http.NewRequest("POST", u, b)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		rsp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error flushing: %s", err)
		}
		defer rsp.Body.Close()
		b, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return err
		}
		if rsp.Status[0] != '2' {
			return fmt.Errorf("expected status 2xx, got %s: %s", rsp.Status, string(b))
		}
		return nil
	}()
	if err != nil {
		log.Printf("err=%q", err)
	}

	i.lastFlush = time.Now()
}

type indexDoc struct {
	Doc *Doc `json:"index"`
}

// add should NOT be called directly as it is not thread safe
// use the channel to add new documents
func (i *BatchIndexer) addInternal(doc *Doc) (int, error) {
	b, err := json.Marshal(&indexDoc{Doc: doc})
	if err != nil {
		return i.cnt, err
	}
	i.buf.Write(b)
	i.buf.Write([]byte("\n"))
	b, err = json.Marshal(doc.Source)
	if err != nil {
		return i.cnt, nil
	}
	i.buf.Write(b)
	i.buf.Write([]byte("\n"))
	i.cnt++
	return i.cnt, nil
}
