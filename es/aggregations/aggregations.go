package aggregations

import (
	"encoding/json"
	"fmt"
)

type Aggregations map[string]*Aggregate

// UnmarshalJSON implement the json.Unmarshaler interface
func (a Aggregations) UnmarshalJSON(b []byte) error {
	var i map[string]interface{}
	e := json.Unmarshal(b, &i)
	if e != nil {
		return e
	}
	if a == nil {
		return fmt.Errorf("Aggregations must be set before unmarshalling")
	}
	return a.load(i)
}

// load initialized the Aggregations map from a generic map
func (a Aggregations) load(i map[string]interface{}) error {
	for k, value := range i {
		valueMap, ok := value.(map[string]interface{})
		if ok {
			agg := &Aggregate{Name: k}
			e := agg.Load(valueMap)
			if e != nil {
				return e
			}
			a[k] = agg
		}
	}
	return nil
}

func loadAggregate(m map[string]interface{}) (i interface{}, e error) {
	i, e = loadStatsAggregate(m)
	if e == nil {
		return i, nil
	}
	i, e = loadBuckets(m)
	if e == nil {
		return i, nil
	}
	i, e = loadPercentileAggregate(m)
	if e == nil {
		return i, nil
	}
	i, e = loadValueAggregate(m)
	if e == nil {
		return i, nil
	}
	return nil, fmt.Errorf("unable to load aggregate from %#v", m)
}

func loadBucket(m map[string]interface{}) (*Bucket, error) {
	b := &Bucket{
		Aggregations: Aggregations{},
	}
	e := b.load(m)
	return b, e
}

func loadBuckets(m map[string]interface{}) (Buckets, error) {
	out := []*Bucket{}
	if len(m) != 1 {
		return nil, fmt.Errorf("not a buckets response 1")
	}
	raw, ok := m["buckets"]
	if !ok {
		return nil, fmt.Errorf("not a buckets response 2")
	}
	buckets, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("not a buckets response 3")
	}

	for _, b := range buckets {
		m, ok := b.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("not a buckets response 3")
		}
		bucket, e := loadBucket(m)
		if e != nil {
			return nil, e
		}
		out = append(out, bucket)
	}
	return out, nil
}
