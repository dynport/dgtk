package aggregations

import "fmt"

type Stats struct {
	Count float64
	Min   float64
	Max   float64
	Avg   float64
	Sum   float64
}

func loadStatsAggregate(i map[string]interface{}) (*Stats, error) {
	count, countOK := readFloat(i, "count")
	min, minOK := readFloat(i, "min")
	max, maxOK := readFloat(i, "max")
	avg, avgOK := readFloat(i, "avg")
	sum, sumOK := readFloat(i, "sum")
	if countOK && minOK && maxOK && avgOK && sumOK {
		return &Stats{
			Count: count,
			Min:   min,
			Max:   max,
			Avg:   avg,
			Sum:   sum,
		}, nil
	}
	return nil, fmt.Errorf("unable to cast to StatsAggregate")
}
