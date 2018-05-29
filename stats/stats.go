package stats

import (
	"fmt"
	"math"
	"sort"
)

type Stats struct {
	Values []int
	sorted bool
	sum    int
}

func (stats *Stats) Add(i int) {
	stats.sum += i
	stats.Values = append(stats.Values, i)
	stats.sorted = false
}

func (stats *Stats) Max() int {
	stats.Sort()
	if len(stats.Values) == 0 {
		return 0
	}
	return stats.Values[len(stats.Values)-1]
}

func (stats *Stats) Min() int {
	stats.Sort()
	if len(stats.Values) == 0 {
		return 0
	}
	return stats.Values[0]
}

func (stats *Stats) Sum() int {
	return stats.sum
}

func (stats *Stats) Sort() {
	if !stats.sorted {
		sort.Ints(stats.Values)
		stats.sorted = true
	}
}

func (stats *Stats) Avg() int {
	return int(math.Floor(float64(stats.Sum()) / float64(len(stats.Values))))
}

func (stats *Stats) Median() int {
	return stats.Perc(50)
}

func Percentile(values []int, perc int) (o int) {
	middle := float64(len(values)) * float64(perc) / 100.0
	floor := int(math.Floor(middle))
	if len(values) <= floor {
		panic(fmt.Sprintf("unabel to get idx %d of %v", floor, values))
	}
	return values[floor]
}

func PercentileFloat(values []float64, perc int) (o float64) {
	middle := float64(len(values)) * float64(perc) / 100.0
	floor := int(math.Floor(middle))
	if len(values) <= floor {
		panic(fmt.Sprintf("unabel to get idx %d of %v", floor, values))
	}
	return values[floor]
}

// Perc calculates the percentile, use 50 for median
func (stats *Stats) Perc(perc int) int {
	stats.Sort()
	return Percentile(stats.Values, perc)
}

func (stats *Stats) String() string {
	return fmt.Sprintf("len: %d, avg: %d, med: %d, perc_95: %d, perc_99: %d, max: %d, min: %d",
		len(stats.Values), stats.Avg(), stats.Median(), stats.Perc(95), stats.Perc(99),
		stats.Max(),
		stats.Min(),
	)
}

func (stats *Stats) Len() int {
	return len(stats.Values)
}

func (stats *Stats) Reset() {
	stats.Values = stats.Values[:0]
	stats.sorted = false
}
