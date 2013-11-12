package es

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	OK  = '2'
	AND = "AND"
	OR  = "OR"
)

type Term map[string]interface{}

type Terms map[string][]interface{}

type Filter struct {
	And   []*Filter         `json:"and,omitempty`
	Term  *Term             `json:"term,omitempty"`
	Terms *Terms            `json:"term,omitempty"`
	Range map[string]*Range `json:"range,omitempty"`
}

type Range struct {
	From interface{}
	To   interface{}
}

type Filtered struct {
	Filter *Filter
}

type QueryString struct {
	Query           string `json:"query,omitempty"`
	DefaultOperator string `json:"default_operator,omitempty"`
}

type Query struct {
	Filtered    *Filtered    `json:"filtered,omitempty"`
	QueryString *QueryString `json:"query_string,omitempty"`
}

var query = Query{
	Filtered: &Filtered{
		Filter: &Filter{
			And: []*Filter{
				{
					Term: &Term{
						"Device": "Anrdoi",
					},
				},
				{
					Terms: &Terms{
						"Action": []interface{}{
							"api/v1/my/photos#create",
							"api/v1/photos#create",
						},
					},
					Range: map[string]*Range{
						"Time": {
							From: "",
						},
					},
				},
			},
		},
	},
}

type BulkIndexJob struct {
	Key    string
	Record interface{}
}

type Index struct {
	Host          string
	Port          int
	Index         string
	Type          string
	bulkIndexJobs []*BulkIndexJob
	BatchSize     int
	Debug         bool
}

func (index *Index) EnqueueBulkIndex(key string, record interface{}) (bool, error) {
	if index.BatchSize == 0 {
		index.BatchSize = 100
	}
	if cap(index.bulkIndexJobs) == 0 {
		index.ResetQueue()
	}
	index.bulkIndexJobs = append(index.bulkIndexJobs, &BulkIndexJob{
		Key: key, Record: record,
	})
	if len(index.bulkIndexJobs) >= index.BatchSize {
		return true, index.RunBatchIndex()
	}
	return false, nil
}

func (index *Index) RunBatchIndex() error {
	started := time.Now()
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	for _, r := range index.bulkIndexJobs {
		enc.Encode(map[string]map[string]string{
			"index": map[string]string{
				"_index": index.Index,
				"_type":  index.Type,
				"_id":    r.Key,
			},
		})
		enc.Encode(r.Record)
	}
	rsp, e := http.Post(index.BaseUrl()+"/_bulk", "application/json", buf)
	if e != nil {
		return e
	}
	defer rsp.Body.Close()
	b, _ := ioutil.ReadAll(rsp.Body)
	if rsp.Status[0] != OK {
		return fmt.Errorf("Error sending bulk request: %s %s", rsp.Status, string(b))
	}
	perSecond := float64(len(index.bulkIndexJobs)) / time.Now().Sub(started).Seconds()
	if index.Debug {
		fmt.Printf("indexed %d, %.1f/second\n", len(index.bulkIndexJobs), perSecond)
	}
	index.ResetQueue()
	return nil
}

func (index *Index) ResetQueue() {
	index.bulkIndexJobs = make([]*BulkIndexJob, 0, index.BatchSize)
}

type IndexStatus struct {
	Index  string `json:"_index"`
	Type   string `json:"_type"`
	Id     string `json:"_id"`
	Exists bool   `json:"exists"`
}

func (index *Index) Status() (status *IndexStatus, e error) {
	rsp, e := http.Get(index.BaseUrl() + "/_status")
	if e != nil {
		return nil, e
	}
	defer rsp.Body.Close()
	b, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return nil, e
	}
	status = &IndexStatus{}
	e = json.Unmarshal(b, status)
	return status, e
}

func (index *Index) Mapping() (i interface{}, e error) {
	rsp, e := index.request("GET", index.IndexUrl()+"/_mapping", i)
	if e != nil {
		if rsp.StatusCode == 404 {
			return nil, nil
		}
		return nil, e
	}
	e = json.Unmarshal(rsp.Body, &i)
	if e != nil {
		return nil, e
	}
	return i, nil
}

func (index *Index) PutMapping(mapping interface{}) (rsp *HttpResponse, e error) {
	return index.request("PUT", index.IndexUrl()+"/", mapping)
}

func (index *Index) BaseUrl() string {
	if index.Port == 0 {
		index.Port = 9200
	}
	return fmt.Sprintf("http://%s:%d", index.Host, index.Port)
}

func (index *Index) IndexUrl() string {
	if index.Index != "" {
		return index.BaseUrl() + "/" + index.Index
	}
	return ""
}

func (index *Index) TypeUrl() string {
	if base := index.IndexUrl(); base != "" && index.Type != "" {
		return base + "/" + index.Type
	}
	return ""
	return ""
}

func (index *Index) Search(req *Request) (rsp *Response, e error) {
	writer := &bytes.Buffer{}
	js := json.NewEncoder(writer)
	e = js.Encode(req)
	if e != nil {
		return nil, e
	}
	u := index.TypeUrl()
	if !strings.HasSuffix(u, "/") {
		u += "/"
	}
	u += "/_search"
	httpRequest, e := http.NewRequest("POST", u, writer)
	if e != nil {
		return nil, e
	}
	httpResponse, e := http.DefaultClient.Do(httpRequest)
	if e != nil {
		return nil, e
	}
	defer httpResponse.Body.Close()
	dec := json.NewDecoder(httpResponse.Body)
	rsp = &Response{}
	e = dec.Decode(rsp)
	if e != nil {
		return nil, e
	}
	return rsp, nil
}

func (index *Index) Post(u string, i interface{}) (*HttpResponse, error) {
	return index.request("POST", u, i)
}

func (index *Index) PutObject(id string, i interface{}) (*HttpResponse, error) {
	return index.request("PUT", index.TypeUrl()+"/"+id, i)
}

func (index *Index) Put(u string, i interface{}) (*HttpResponse, error) {
	return index.request("PUT", u, i)
}

type HttpResponse struct {
	*http.Response
	Body []byte
}

func (index *Index) request(method string, u string, i interface{}) (httpResponse *HttpResponse, e error) {
	var req *http.Request
	if i != nil {
		buf := &bytes.Buffer{}
		encoder := json.NewEncoder(buf)
		if e := encoder.Encode(i); e != nil {
			return nil, e
		}
		req, e = http.NewRequest(method, u, buf)
	} else {
		req, e = http.NewRequest(method, u, nil)
	}
	if e != nil {
		return nil, e
	}
	rsp, e := http.DefaultClient.Do(req)
	if e != nil {
		return nil, e
	}
	defer rsp.Body.Close()
	b, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return nil, e
	}
	httpResponse = &HttpResponse{
		Response: rsp,
		Body:     b,
	}
	if e != nil {
		return httpResponse, e
	}
	if rsp.Status[0] != OK {
		return httpResponse, fmt.Errorf("error indexing: %s %s", rsp.Status, string(b))
	}
	return httpResponse, nil
}
