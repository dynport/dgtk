package opentsdb

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Logger interface {
	Debug(i ...interface{})
}

type DefaultLogger struct {
}

func (logger *DefaultLogger) Debug(i ...interface{}) {
	return
}

func SetLogger(newLogger Logger) {
	logger = newLogger
}

var logger Logger

func init() {
	logger = &DefaultLogger{}
}

// OpenTSDB request parameters.
type RequestParams struct {
	Host    string                 // Host to query.
	Start   string                 // Time point when to start query.
	End     string                 // Time point to end query (optional).
	Metrics []*MetricConfiguration // Configuration of the metrics to request.
}

// OpenTSDB metric query parameters and configuration for result
// interpration.
type MetricConfiguration struct {
	Unit      string                // TODO: required?
	Filter    func(float64) float64 // Function used to map metric values.
	Aggregate string                // Aggregation of matching metrics
	Rate      string                // Mark metric as rate or downsample.
	Metric    string                // Metric to query for.
	TagFilter string                // Filter on tags (comma separated string with <tag>=<value> pairs.
}

// Mapping from the metric identifier to the according configuration
// used to parse and handle the results.
type MetricConfigurations map[string]*MetricConfiguration

// Parse a single line of the result returned by OpenTSDB in ASCII mode.
func parseLogEventLine(line string, mCfg MetricConfigurations) (*MetricValue, error) {
	mv := &MetricValue{}
	if e := mv.Parse(line); e != nil {
		return nil, e
	}
	if mCfg[mv.Key].Filter != nil {
		mv.Value = mCfg[mv.Key].Filter(mv.Value)
	}
	return mv, nil
}

// Parse the content of the ASCII based OpenTSDB response.
func parseResponse(content io.ReadCloser, mCfg MetricConfigurations) (MetricsTree, error) {
	scanner := bufio.NewScanner(content)
	mt := NewMetricsTree()
	cnt := 0
	dur := map[string]time.Duration{}
	started := time.Now()
	for {
		cnt++
		started := time.Now()
		if !scanner.Scan() {
			break
		}
		l := scanner.Text()
		dur["read"] += time.Now().Sub(started)
		started = time.Now()
		if mv, e := parseLogEventLine(l, mCfg); e == nil {
			if e = mt.AddMetricValue(mv); e != nil {
				return nil, e
			}
		} else {
			return nil, e
		}
		dur["parse"] += time.Now().Sub(started)
	}
	logger.Debug(fmt.Sprintf("parsed %d in %s (%v)", cnt, time.Now().Sub(started), dur))
	return mt, nil
}

func createQueryURL(attrs *RequestParams) string {
	values := url.Values{}
	values.Add("start", attrs.Start)
	if attrs.End != "" {
		values.Add("end", attrs.End)
	}

	for _, m := range attrs.Metrics {
		metric := m.Aggregate
		if m.Rate != "" {
			metric += ":" + m.Rate
		}
		metric += ":" + m.Metric
		metric += "{" + m.TagFilter + "}"
		values.Add("m", metric)
	}

	return "http://" + attrs.Host + ":4242/q?ascii&" + values.Encode()
}

func createMetricConfigurations(attrs *RequestParams) (MetricConfigurations, error) {
	mCfg := make(MetricConfigurations)

	for _, m := range attrs.Metrics {
		if _, ok := mCfg[m.Metric]; ok {
			return nil, errors.New("Each metric only allowed once!")
		}
		mCfg[m.Metric] = m
	}
	return mCfg, nil
}

// Request data from OpenTSDB in ASCII format.
func GetData(attrs *RequestParams) (MetricsTree, error) {
	url := createQueryURL(attrs)
	logger.Debug("Request URL is ", url)

	mCfg, err := createMetricConfigurations(attrs)
	if err != nil {
		return nil, err
	}

	logger.Debug("Starting request to OpenTSDB: " + url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(fmt.Sprintf("Request to OpenTSDB failed with %s (%s)", resp.Status, string(b)))
	}
	logger.Debug("Finished request to OpenTSDB")

	logger.Debug("Starting to parse the response from OpenTSDB")
	started := time.Now()
	mt, e := parseResponse(resp.Body, mCfg)
	logger.Debug(fmt.Sprintf("Finished parsing the response from OpenTSDB in %s", time.Now().Sub(started)))

	return mt, e
}
