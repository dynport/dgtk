package aggregations

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/dynport/dgtk/tskip/tskip"
)

func mustReadFixture(t *testing.T, name string) []byte {
	b, e := ioutil.ReadFile("fixtures/" + name)
	if e != nil {
		tskip.Errorf(t, 1, "error reading fixture %s", e)
		t.FailNow()
	}
	return b
}

func TestMarshallAggregations(t *testing.T) {
	b := mustReadFixture(t, "response_with_aggregations.json")

	type req struct {
		Aggs json.RawMessage `json:"aggregations"`
	}

	r := &req{}
	failIfError(t, json.Unmarshal(b, r))

	b = r.Aggs

	v := Aggregations{}

	failIfError(t, json.Unmarshal(b, &v))

	assertEqual(t, len(v), 1)

	bucket := v["days"].Buckets[0]
	assertEqual(t, bucket.DocCount, 2341)

	aggs := bucket.Aggregations
	assertEqual(t, len(aggs), 2)

	types, ok := aggs["types"]
	if !ok {
		t.Fatal("types aggregation must be found")
	}
	failIfNil(t, types)
	uniques, ok := aggs["uniques"]
	if !ok {
		t.Fatal("uniques aggregation must be found")
	}
	failIfNil(t, uniques)
	failIfNil(t, uniques.Value)
	assertEqual(t, uniques.Value.Value, 1991.0)

	assertEqual(t, len(types.Buckets), 5)

	bucket = types.Buckets[0]
	assertEqual(t, fmt.Sprint(bucket.Key), "1")
}

func TestUnmarshalAggregations(t *testing.T) {
	b := mustReadFixture(t, "aggregations.json")

	agg := Aggregations{}
	failIfError(t, json.Unmarshal(b, &agg))
	assertEqual(t, len(agg), 1)
}

func TestLoadBucket(t *testing.T) {
	raw := []byte(`{"key": 1, "doc_count": 2325}`)
	b := &Bucket{}
	failIfError(t, json.Unmarshal(raw, b))
	assertEqual(t, b.Key, 1.0)
	assertEqual(t, b.DocCount, 2325)
	assertEqual(t, b.KeyAsString, "")

	raw = []byte(`{"key": 1, "doc_count": 2325, "key_as_string": "test"}`)
	b = &Bucket{}
	failIfError(t, json.Unmarshal(raw, b))
	assertEqual(t, b.Key, 1.0)
	assertEqual(t, b.DocCount, 2325)
	assertEqual(t, b.KeyAsString, "test")
}

func TestLoadStatsAggregate(t *testing.T) {
	//raw := []byte(`{"count": 609, "min": 3, "max": 1978, "avg": 99.2495894909688, "sum": 60443}`)
	m := map[string]interface{}{
		"count": 609.0,
		"min":   3.0,
		"max":   1978.0,
		"avg":   99.2495894909688,
		"sum":   60443.0,
	}
	agg, e := loadStatsAggregate(m)
	failIfError(t, e)
	assertEqual(t, agg.Count, 609.0)
}

func TestLoadValueAggregate(t *testing.T) {

	m := map[string]interface{}{
		"value": 10.0,
	}
	agg, e := loadValueAggregate(m)
	failIfError(t, e)
	assertEqual(t, agg.Value, 10.0)
}

func TestLoadPercentileAggregate(t *testing.T) {
	m := map[string]interface{}{
		"1.0": 6.0,
		"5.0": 8.0,
	}
	agg, e := loadPercentileAggregate(m)
	failIfError(t, e)
	assertEqual(t, agg.Values[1.0], 6.0)
	assertEqual(t, agg.Values[5.0], 8.0)
}

func TestLoadBucket2(t *testing.T) {
	m := map[string]interface{}{
		"key_as_string": "2014-05-11T00:00:00.000Z",
		"key":           1399766400000,
		"doc_count":     2341.00,
	}
	bucket, e := loadBucket(m)
	failIfError(t, e)
	assertEqual(t, bucket.DocCount, 2341)
}

func TestLoadBuckets(t *testing.T) {
	m := map[string]interface{}{
		"buckets": []interface{}{
			map[string]interface{}{
				"key_as_string": "2014-05-11T00:00:00.000Z",
				"key":           1399766400000,
				"doc_count":     2341.0,
			},
		},
	}
	agg, e := loadBuckets(m)
	failIfError(t, e)
	assertEqual(t, len(agg), 1)
	first := agg[0]
	failIfNil(t, first)
	assertEqual(t, first.DocCount, 2341)
	assertEqual(t, first.KeyAsString, "2014-05-11T00:00:00.000Z")
}
