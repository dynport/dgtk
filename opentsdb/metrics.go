package opentsdb

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Data type for a metric value.
type MetricValue struct {
	Key   string    // Metric's key.
	Value float64   // Metric's value.
	Time  time.Time // Timestamp of metric recording.
	Tags  string    // Tags the metric has set.
}

func (mv *MetricValue) String() string {
	return fmt.Sprintf("%s %s %.01f %s", mv.Time.Format("2006-01-02T15:04:05"), mv.Key, mv.Value, mv.Tags)
}

// Aggregation of metric values.
type MetricsAggregate struct {
	firstMetric time.Time // Timestamp of metric recorded first.
	lastMetric  time.Time // Timestamp of metric recorded last.
	Sum         float64   // Sum of all values recorded.
	Avg         float64   // Average value recorded.
	Min         float64   // Minimum value recorded.
	Max         float64   // Maximum value recorded.
}

func (aggregate *MetricsAggregate) Diff() float64 {
	return aggregate.Max - aggregate.Min
}

func (aggregate *MetricsAggregate) Duration() time.Duration {
	return aggregate.lastMetric.Sub(aggregate.firstMetric)
}

// List of metric values.
type MetricsList struct {
	data      []*MetricValue    // Raw data returned from OpenTSDB.
	aggregate *MetricsAggregate // Cache for metrics aggregates.
	tags      map[string]string // Cache for tag values.
	isSorted  bool              // Flag whether data is sorted or not.
}

func (list *MetricsList) Data() []*MetricValue {
	return list.data
}

// Create a new metrics list for the given tags.
func NewMetricsList(tags string) (*MetricsList, error) {
	tagMap := make(map[string]string)
	for _, element := range strings.Split(tags, " ") {
		parts := strings.SplitN(element, "=", 2)
		if len(parts) == 2 {
			key, value := parts[0], parts[1]
			tagMap[key] = value
		} else {
			return nil, errors.New(fmt.Sprintf("failed to parse tags: '%s'", tags))
		}
	}
	return &MetricsList{tags: tagMap, isSorted: false}, nil
}

func (mvl *MetricsList) Len() int {
	return len(mvl.data)
}

func (mvl *MetricsList) Less(a, b int) bool {
	return mvl.data[a].Time.Sub(mvl.data[b].Time) < 0
}

func (mvl *MetricsList) Swap(a, b int) {
	mvl.data[a], mvl.data[b] = mvl.data[b], mvl.data[a]
}

func (mvl *MetricsList) Add(mv *MetricValue) {
	mvl.aggregate = nil
	mvl.isSorted = false
	mvl.data = append(mvl.data, mv)
}

func (mvl *MetricsList) Sort() {
	if !mvl.isSorted {
		sort.Sort(mvl)
		mvl.isSorted = true
	}
}

// Returns the first metric recorded.
func (mvl *MetricsList) First() *MetricValue {
	if len(mvl.data) > 0 {
		mvl.Sort()
		return mvl.data[0]
	}
	return nil
}

// Returns the last metric recorded.
func (mvl *MetricsList) Last() *MetricValue {
	if len(mvl.data) > 0 {
		mvl.Sort()
		return mvl.data[len(mvl.data)-1]
	}
	return nil
}

// Returns aggregate values on the list of data. The computations are
// cached and returned directly if available. Cache will be
// invalidated if another value is added.
func (mvl *MetricsList) Aggregate() (*MetricsAggregate, error) {
	if len(mvl.data) == 0 {
		return nil, fmt.Errorf("MetricsList contains no data!")
	}

	if mvl.aggregate != nil {
		return mvl.aggregate, nil
	}

	mvl.Sort()
	sum, min, max := 0.0, mvl.First().Value, mvl.First().Value

	for _, v := range mvl.data {
		sum += v.Value
		switch {
		case v.Value < min:
			min = v.Value
		case v.Value > max:
			max = v.Value
		}
	}

	mvl.aggregate = &MetricsAggregate{
		Sum:         sum,
		firstMetric: mvl.First().Time,
		lastMetric:  mvl.Last().Time,
		Min:         min,
		Max:         max,
		Avg:         sum / float64(len(mvl.data)),
	}

	return mvl.aggregate, nil
}

// Get value of a given tag.
func (mvl *MetricsList) GetTagValue(tag string) (value string, e error) {
	if value, ok := mvl.tags[tag]; !ok {
		return value, errors.New("Tag unknown!")
	} else {
		return value, e
	}
}

// The metrics tree is a data structure containing metric values
// sorted after keys and tags.
type MetricsTree map[string]tagMap
type tagMap map[string]*MetricsList

func NewMetricsTree() MetricsTree {
	return make(MetricsTree)
}

// Add a new metric value to the metrics tree.
func (mt MetricsTree) AddMetricValue(mv *MetricValue) error {
	tm, found := mt[mv.Key]
	if !found {
		tm = make(tagMap)
		mt[mv.Key] = tm
	}
	metricsList, found := tm[mv.Tags]
	if !found {
		var e error
		metricsList, e = NewMetricsList(mv.Tags)
		if e != nil {
			return e
		}
		tm[mv.Tags] = metricsList
	}
	metricsList.Add(mv)
	return nil
}

func (mt MetricsTree) GetMetricsList(key, tags string) (*MetricsList, error) {
	if _, ok := mt[key]; !ok {
		return nil, fmt.Errorf("Key '%s' does not exist!", key)
	}
	if _, ok := mt[key][tags]; !ok {
		return nil, fmt.Errorf("Tags '%s' do not exist for key '%s'!", tags, key)
	}
	return mt[key][tags], nil
}

// Get all tags of the given metric that are returned by the according
// query.
func (mt MetricsTree) GetTagsForMetric(metric string) (tags []string, e error) {
	if m, ok := mt[metric]; ok {
		keys := make([]string, 0, len(m))
		for key, _ := range m {
			keys = append(keys, key)
		}
		return keys, e
	}
	return tags, errors.New(fmt.Sprintf("Failed to find metric %s", metric))
}

func (mt MetricsTree) GetMetricsListAggregate(metric, tag string) (ma *MetricsAggregate, e error) {
	var ml *MetricsList
	if ml, e = mt.GetMetricsList(metric, tag); e == nil {
		if ma, e = ml.Aggregate(); e == nil {
			return ma, e
		}
	}
	return ma, e
}
