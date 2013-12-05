package es

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	OK  = '2'
	AND = "AND"
	OR  = "OR"
)

type Analysis struct {
	Analyzer Analyzer `json:"analyzer"`
}

type Analyzer struct {
	Default AnalyzerType `json:"default"`
}

type AnalyzerType struct {
	Type string `json:"type"`
}

type IndexConfig struct {
	Index IndexIndexConfig `json:"index"`
}

type IndexIndexConfig struct {
	Analysis Analysis `json:"analysis"`
}

func KeywordIndex() IndexConfig {
	return IndexConfig{
		Index: IndexIndexConfig{
			Analysis: Analysis{
				Analyzer: Analyzer{
					Default: AnalyzerType{
						Type: "keyword",
					},
				},
			},
		},
	}
}

func (index *Index) CreateIndex(config IndexConfig) (rsp *HttpResponse, e error) {
	return index.request("PUT", index.IndexUrl(), config)
}

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

func (index *Index) IndexExists() (exists bool, e error) {
	if index.Index == "" {
		return false, fmt.Errorf("no index set")
	}
	rsp, e := http.Get(index.IndexUrl() + "/_status")
	if e != nil {
		return false, e
	}
	return rsp.Status[0] == '2', nil
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

func (index *Index) DeleteIndex() error {
	req, e := http.NewRequest("DELETE", index.IndexUrl(), nil)
	if e != nil {
		return e
	}
	rsp, e := http.DefaultClient.Do(req)
	if e != nil {
		return e
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		return fmt.Errorf("Error delting index at %s: %s", index.TypeUrl(), rsp.Status)
	}
	return nil
}

func (index *Index) Refresh() error {
	rsp, e := index.request("POST", index.IndexUrl()+"/_refresh", nil)
	if e != nil {
		return e
	}
	if rsp.Status[0] != '2' {
		return fmt.Errorf("Error refreshing index: %s", rsp.Status)
	}
	return nil
}

func (index *Index) RunBatchIndex() error {
	if len(index.bulkIndexJobs) == 0 {
		return nil
	}
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

func (index *Index) GlobalMapping() (m *Mapping, e error) {
	m = &Mapping{}
	u := index.BaseUrl() + "/_mapping"
	log.Printf("checking for url %s", u)
	rsp, e := index.request("GET", u, m)
	if rsp != nil && rsp.StatusCode == 404 {
		return nil, nil
	} else if e != nil {
		return nil, e
	}
	e = json.Unmarshal(rsp.Body, m)
	if e != nil {
		return nil, e
	}
	return m, nil
}

func (index *Index) Mapping() (i interface{}, e error) {
	u := index.IndexUrl() + "/_mapping"
	log.Printf("checking for url %s", u)
	rsp, e := index.request("GET", u, i)
	if rsp != nil && rsp.StatusCode == 404 {
		return nil, nil
	} else if e != nil {
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
	if index.Type == "" {
		return index.IndexUrl()
	} else if idx := index.IndexUrl(); idx != "" {
		return idx + "/" + index.Type
	} else {
		return ""
	}
}

func (index *Index) DeleteByQuery(query string) error {
	q := &url.Values{}
	q.Add("q", query)
	rsp, e := index.request("DELETE", index.TypeUrl()+"/_query?"+q.Encode(), nil)
	if e != nil {
		return e
	}
	if rsp.Status[0] != '2' {
		return fmt.Errorf("error deleting with query: %s", rsp.Status)
	}
	return nil
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

func (index *Index) PostObject(i interface{}) (*HttpResponse, error) {
	return index.request("POST", index.TypeUrl(), i)
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
