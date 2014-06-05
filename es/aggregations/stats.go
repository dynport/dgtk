package aggregations

import (
	"encoding/json"
	"fmt"
)

type Stats struct {
	Field        string                    `json:"field"`
	Aggregations map[string]json.Marshaler `json:"aggregations,omitempty"`
}

func (a *Stats) MarshalJSON() ([]byte, error) {
	h := hash{
		"stats": hash{
			"field": a.Field,
		},
	}
	if a.Aggregations != nil {
		h["aggregations"] = a.Aggregations
	}

	return json.Marshal(h)
}

type StatsAggregate struct {
	Count float64
	Min   float64
	Max   float64
	Avg   float64
	Sum   float64
}

func loadStatsAggregate(i map[string]interface{}) (*StatsAggregate, error) {
	count, countOK := readFloat(i, "count")
	min, minOK := readFloat(i, "min")
	max, maxOK := readFloat(i, "max")
	avg, avgOK := readFloat(i, "avg")
	sum, sumOK := readFloat(i, "sum")
	if countOK && minOK && maxOK && avgOK && sumOK {
		return &StatsAggregate{
			Count: count,
			Min:   min,
			Max:   max,
			Avg:   avg,
			Sum:   sum,
		}, nil
	}
	return nil, fmt.Errorf("unable to cast to StatsAggregate")
}
