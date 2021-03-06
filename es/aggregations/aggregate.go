package aggregations

import "fmt"

type Aggregate struct {
	Name       string          `json:"name,omitempty"`
	Stats      *StatsAggregate `json:"stats,omitempty"`
	Value      *Value          `json:"value,omitempty"`
	Percentile *Percentile     `json:"percentile,omitempty"`
	Buckets    Buckets         `json:"buckets,omitempty"`
}

func (agg *Aggregate) Load(m map[string]interface{}) error {
	raw, e := loadAggregate(m)
	if e != nil {
		return e
	}
	switch a := raw.(type) {
	case Buckets:
		agg.Buckets = a
	case *StatsAggregate:
		agg.Stats = a
	case *Value:
		agg.Value = a
	case *Percentile:
		agg.Percentile = a
	default:
		return fmt.Errorf("unable to map %#v (%T) to Aggregate", a, a)
	}
	return nil
}
