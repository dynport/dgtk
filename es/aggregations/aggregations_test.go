package aggregations

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func mustReadFixture(t *testing.T, name string) []byte {
	b, e := ioutil.ReadFile("fixtures/" + name)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

func TestMarshalling(t *testing.T) {
	Convey("marshal aggregations", t, func() {
		So(1, ShouldEqual, 1)
	})
}

func TestMarshallAggregations(t *testing.T) {
	Convey("Marshal Aggs", t, func() {
		b := mustReadFixture(t, "response_with_aggregations.json")
		So(b, ShouldNotBeNil)

		type req struct {
			Aggs json.RawMessage `json:"aggregations"`
		}

		r := &req{}
		So(json.Unmarshal(b, r), ShouldBeNil)

		b = r.Aggs

		v := Aggregations{}

		e := json.Unmarshal(b, &v)
		So(e, ShouldBeNil)

		So(len(v), ShouldEqual, 1)

		bucket := v["days"].Buckets[0]
		So(bucket.DocCount, ShouldEqual, 2341)

		aggs := bucket.Aggregations
		So(len(aggs), ShouldEqual, 2)

		types, ok := aggs["types"]
		So(ok, ShouldEqual, true)
		So(types, ShouldNotBeNil)

		uniques, ok := aggs["uniques"]
		So(ok, ShouldEqual, true)
		So(uniques, ShouldNotBeNil)
		So(uniques.Value, ShouldNotBeNil)
		So(uniques.Value.Value, ShouldEqual, 1991)

		So(len(types.Buckets), ShouldEqual, 5)

		bucket = types.Buckets[0]
		So(fmt.Sprint(bucket.Key), ShouldEqual, "1")
	})

	Convey("Unmarshal", t, func() {
		b := mustReadFixture(t, "aggregations.json")
		So(b, ShouldNotBeNil)

		agg := Aggregations{}
		e := json.Unmarshal(b, &agg)
		So(e, ShouldBeNil)
		So(len(agg), ShouldEqual, 1)
	})

	Convey("Load Bucket", t, func() {
		raw := []byte(`{"key": 1, "doc_count": 2325}`)
		b := &Bucket{}
		So(json.Unmarshal(raw, b), ShouldBeNil)
		So(b.Key, ShouldEqual, 1)
		So(b.DocCount, ShouldEqual, 2325)
		So(b.KeyAsString, ShouldEqual, "")

		raw = []byte(`{"key": 1, "doc_count": 2325, "key_as_string": "test"}`)
		b = &Bucket{}
		So(json.Unmarshal(raw, b), ShouldBeNil)
		So(b.Key, ShouldEqual, 1)
		So(b.DocCount, ShouldEqual, 2325)
		So(b.KeyAsString, ShouldEqual, "test")
	})

	Convey("Load stats aggregate", t, func() {
		//raw := []byte(`{"count": 609, "min": 3, "max": 1978, "avg": 99.2495894909688, "sum": 60443}`)
		m := map[string]interface{}{
			"count": 609.0,
			"min":   3.0,
			"max":   1978.0,
			"avg":   99.2495894909688,
			"sum":   60443.0,
		}
		agg, e := loadStatsAggregate(m)
		So(e, ShouldBeNil)
		So(agg, ShouldNotBeNil)
		So(agg.Count, ShouldEqual, 609.0)
	})

	Convey("load value aggregate", t, func() {
		m := map[string]interface{}{
			"value": 10.0,
		}
		agg, e := loadValueAggregate(m)
		So(e, ShouldBeNil)
		So(agg, ShouldNotBeNil)
		So(agg.Value, ShouldEqual, 10.0)
	})

	Convey("Load percentile aggregate", t, func() {
		m := map[string]interface{}{
			"1.0": 6.0,
			"5.0": 8.0,
		}
		agg, e := loadPercentileAggregate(m)
		So(e, ShouldBeNil)
		So(agg, ShouldNotBeNil)
		So(agg.Values[1.0], ShouldEqual, 6.0)
		So(agg.Values[5.0], ShouldEqual, 8.0)
	})

	Convey("load bucket", t, func() {
		m := map[string]interface{}{
			"key_as_string": "2014-05-11T00:00:00.000Z",
			"key":           1399766400000,
			"doc_count":     2341.00,
		}
		bucket, e := loadBucket(m)
		So(e, ShouldBeNil)
		So(bucket.DocCount, ShouldEqual, 2341)
	})

	Convey("load buckets", t, func() {
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
		So(e, ShouldBeNil)
		So(len(agg), ShouldEqual, 1)
		first := agg[0]
		So(first, ShouldNotBeNil)
		So(first.DocCount, ShouldEqual, 2341)
		So(first.KeyAsString, ShouldEqual, "2014-05-11T00:00:00.000Z")
	})
}
