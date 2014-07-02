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

	"github.com/dynport/dgtk/es/aggregations"
)

const (
	OK  = '2'
	AND = "AND"
	OR  = "OR"
)

type Logger interface {
	Debug(format string, i ...interface{})
	Info(format string, i ...interface{})
	Error(format string, i ...interface{})
}

type Analysis struct {
	Analyzer Analyzer `json:"analyzer"`
}

type Analyzer struct {
	Default AnalyzerType `json:"default"`
}

type AnalyzerType struct {
	Type      string `json:"type,omitempty"`
	Tokenizer string `json:"tokenizer,omitempty"`
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
						Tokenizer: "whitespace",
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
	Not   *Not              `json:"not,omitempty"`
	And   []*Filter         `json:"and,omitempty"`
	Term  *Term             `json:"term,omitempty"`
	Terms *Terms            `json:"term,omitempty"`
	Range map[string]*Range `json:"range,omitempty"`
}

type Not struct {
	Filter *Filter `json:"filter,omitempty"`
}

type Range struct {
	From interface{}
	To   interface{}
}

type Filtered struct {
	Filter *Filter
	Query  *Query
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
	Type   string
	Index  string
	Record interface{}
}

type Index struct {
	Host      string
	Port      int
	Index     string
	Type      string
	batchDocs []*Doc
	BatchSize int
	Debug     bool
	Logger    Logger
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

func (index *Index) EnqueueDoc(doc *Doc) (indexed bool, e error) {
	if index.BatchSize == 0 {
		index.BatchSize = 100
	}
	if cap(index.batchDocs) == 0 {
		index.ResetBatch()
	}
	index.batchDocs = append(index.batchDocs, doc)
	if len(index.batchDocs) >= index.BatchSize {
		return true, index.RunBatchIndex()
	}
	return false, nil
}

func (index *Index) ResetBatch() {
	index.batchDocs = make([]*Doc, 0, index.BatchSize)
}

func (index *Index) EnqueueBulkIndex(key string, record interface{}) (bool, error) {
	return index.EnqueueDoc(&Doc{
		Id: key, Source: record,
	})
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

func (index *Index) IndexDocs(docs []*Doc) error {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	for _, doc := range docs {
		if doc.Index == "" {
			doc.Index = index.Index
		}
		if doc.Type == "" {
			doc.Type = index.Type
		}
		enc.Encode(map[string]map[string]string{
			"index": doc.IndexAttributes(),
		})
		enc.Encode(doc.Source)
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
	return nil
}

func (index *Index) RunBatchIndex() error {
	if len(index.batchDocs) == 0 {
		return nil
	}
	started := time.Now()
	e := index.IndexDocs(index.batchDocs)
	if e != nil {
		return e
	}
	perSecond := float64(len(index.batchDocs)) / time.Now().Sub(started).Seconds()
	index.LogDebug("indexed %d, %.1f/second\n", len(index.batchDocs), perSecond)
	index.ResetBatch()
	return nil
}

type Status struct {
	Ok     bool    `json:"ok"`
	Shards *Shards `json:"_shards"`
}

type Shards struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Failed     int `json:"failed"`
}

func (index *Index) Status() (status *Status, e error) {
	rsp, e := http.Get(index.BaseUrl() + "/_status")
	if e != nil {
		return nil, e
	}
	defer rsp.Body.Close()
	b, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return nil, e
	}
	if rsp.Status[0] != '2' {
		return nil, fmt.Errorf("Status: %d, Response: %s", rsp.StatusCode, string(b))
	}
	status = &Status{}
	e = json.Unmarshal(b, status)
	return status, e
}

func init() {
	log.SetFlags(0)
}

func (index *Index) LogDebug(format string, i ...interface{}) {
	if index.Logger != nil {
		index.Logger.Debug(format, i...)
	}
}

func (index *Index) LogInfo(format string, i ...interface{}) {
	if index.Logger != nil {
		index.Logger.Info(format, i...)
	}
}

func (index *Index) GlobalMapping() (m *Mapping, e error) {
	m = &Mapping{}
	u := index.BaseUrl() + "/_mapping"
	index.LogInfo("checking for url %s", u)
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

func (index *Index) Stats() (*Stats, error) {
	u := index.IndexUrl() + "/_stats"
	dbg.Printf("requesting %q", u)
	rsp, e := http.Get(u)
	if e != nil {
		return nil, e
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		return nil, fmt.Errorf("expected status 2xx, got %s", rsp.Status)
	}
	stats := &Stats{}
	e = json.NewDecoder(rsp.Body).Decode(stats)
	return stats, e
}

func (index *Index) Mapping() (i interface{}, e error) {
	u := index.IndexUrl() + "/_mapping"
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
	if index.Index == "" {
		return index.BaseUrl()
	} else {
		return index.BaseUrl() + "/" + index.Index
	}
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

func (index *Index) DeleteByQuery(query string) (b []byte, e error) {
	q := &url.Values{}
	q.Add("q", query)
	rsp, e := index.request("DELETE", index.TypeUrl()+"/_query?"+q.Encode(), nil)
	if e != nil {
		return b, e
	}
	if rsp.Status[0] != '2' {
		return b, fmt.Errorf("error deleting with query: %s", rsp.Status)
	}
	return rsp.Body, nil
}

func (index *Index) SearchRaw(req interface{}) (*ResponseRaw, error) {
	rsp := &ResponseRaw{Response: &Response{}}
	rsp.Aggregations = aggregations.Aggregations{}
	e := index.loadSearch(req, rsp)
	return rsp, e
}

func (index *Index) Search(req interface{}) (*Response, error) {
	rsp := &Response{}
	rsp.Aggregations = aggregations.Aggregations{}
	e := index.loadSearch(req, rsp)
	return rsp, e
}

type Sharder interface {
	ShardsResponse() *ShardsResponse
}

func (index *Index) loadSearch(req interface{}, rsp Sharder) error {
	u := index.TypeUrl()
	if !strings.HasSuffix(u, "/") {
		u += "/"
	}
	u += "/_search"
	httpRequest, e := http.NewRequest("POST", u, nil)
	if e != nil {
		return e
	}
	if req != nil {
		writer := &bytes.Buffer{}
		e := json.NewEncoder(writer).Encode(req)
		if e != nil {
			return e
		}
		httpRequest.Body = ioutil.NopCloser(writer)
	}
	httpResponse, e := http.DefaultClient.Do(httpRequest)
	if e != nil {
		return e
	}
	defer httpResponse.Body.Close()
	b, e := ioutil.ReadAll(httpResponse.Body)
	if e != nil {
		return e
	}
	dbg.Printf("resonse: %s", string(b))
	if httpResponse.Status[0] != '2' {
		return fmt.Errorf("expected staus 2xx, git %s", httpResponse.Status, string(b))
	}
	e = json.Unmarshal(b, rsp)
	if e != nil {
		return e
	}
	shards := rsp.ShardsResponse()
	if shards != nil && len(shards.Failures) > 0 {
		b, _ := json.Marshal(shards.Failures)
		return fmt.Errorf("%s", string(b))
	}
	return nil
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
