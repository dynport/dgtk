package aggregations

import (
	"fmt"
	"strconv"
)

type Percentile struct {
	Values map[float64]float64
}

func loadPercentileAggregate(i map[string]interface{}) (*Percentile, error) {
	m := map[float64]float64{}
	agg := &Percentile{Values: m}
	for k, raw := range i {
		if value, ok := raw.(float64); ok {
			keyValue, e := strconv.ParseFloat(k, 64)
			if e == nil {
				agg.Values[keyValue] = value
				continue
			}
		}
		return nil, fmt.Errorf("not a PercentileAggregate")
	}
	return agg, nil
}
