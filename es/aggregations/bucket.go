package aggregations

import (
	"encoding/json"
	"fmt"
)

type Buckets []*Bucket

type Bucket struct {
	Key          interface{}
	KeyAsString  string
	DocCount     int
	Aggregations Aggregations
}

func (b *Bucket) load(m map[string]interface{}) error {
	b.Aggregations = map[string]*Aggregate{}
	docCount, docOk := readFloat(m, "doc_count")
	key, keyOk := m["key"]
	if docOk && keyOk {
		b.DocCount = int(docCount)
		b.KeyAsString, _ = m["key_as_string"].(string)
		b.Key = key
		for k, v := range m {
			switch k {
			case "doc_count", "key", "key_as_string":
				// ignore
			default:
				subMap, ok := v.(map[string]interface{})
				if !ok {
					return fmt.Errorf("submap is not a map")
				}
				agg := &Aggregate{Name: k}
				e := agg.Load(subMap)
				if e != nil {
					return e
				}
				b.Aggregations[k] = agg
			}
		}
		return nil
	}
	return fmt.Errorf("unable to parse bucket docs=%t key=%t", docOk, keyOk)
}

func (bucket *Bucket) UnmarshalJSON(b []byte) error {
	var i map[string]interface{}
	e := json.Unmarshal(b, &i)
	if e != nil {
		return e
	}
	return bucket.load(i)
}
