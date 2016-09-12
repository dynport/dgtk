package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type restore struct {
	Address   string `cli:"opt -a default=http://127.0.0.1:9200"`
	BatchSize int    `cli:"opt --batch-size default=1000"`
}

func (a *restore) Run() error {
	addr := a.Address
	r := bufio.NewReader(os.Stdin)
	del := byte('\n')
	buf := &bytes.Buffer{}
	cnt := 0
	for {
		b, err := r.ReadBytes(del)
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		var res *record
		err = json.Unmarshal(b, &res)
		if err != nil {
			return err
		}
		src := res.Source
		res.Source = nil
		b, err = json.Marshal(res)
		if err != nil {
			return err
		}
		_, err = io.WriteString(buf, `{"index":`+string(b)+`}`+"\n"+string(src)+"\n")
		if err != nil {
			return err
		}
		cnt++
		if cnt%a.BatchSize == 0 {
			err = flush(addr, buf)
			if err != nil {
				return err
			}
			buf.Reset()
		}
	}
	if buf.Len() > 0 {
		return flush(addr, buf)
	}
	return nil
}

func flush(addr string, r io.Reader) error {
	rsp, err := http.Post(addr+"/_bulk", "application/json", r)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		b, _ := ioutil.ReadAll(rsp.Body)
		return fmt.Errorf("got status %s but expected 2x. body=%s", rsp.Status, string(b))
	}
	return nil
}

type record struct {
	Index  string          `json:"_index,omitempty"`
	Type   string          `json:"_type,omitempty"`
	ID     string          `json:"_id,omitempty"`
	Source json.RawMessage `json:"_source,omitempty"`
}
