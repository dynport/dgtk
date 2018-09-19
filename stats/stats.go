package stats

import (
	"fmt"
	"math"
	"sort"
)

func New(values ...float64) *Stats {
	s := &Stats{}
	s.Add(values...)
	return s
}

type Stats struct {
	values []float64
	sorted bool
	sum    float64
}

func (stats *Stats) Values() []float64 {
	return stats.values
}

func (stats *Stats) Add(values ...float64) {
	for _, v := range values {
		stats.sum += v
		stats.values = append(stats.values, v)
	}
	stats.sorted = false
}

func (stats *Stats) Max() float64 {
	stats.Sort()
	if len(stats.values) == 0 {
		return 0
	}
	return stats.values[len(stats.values)-1]
}

func (stats *Stats) Min() float64 {
	stats.Sort()
	if len(stats.values) == 0 {
		return 0
	}
	return stats.values[0]
}

func (stats *Stats) Sum() float64 {
	return stats.sum
}

func (stats *Stats) Sort() {
	if !stats.sorted {
		sort.Float64s(stats.values)
		stats.sorted = true
	}
}

func (stats *Stats) Avg() float64 {
	return stats.sum / float64(len(stats.values))
}

func (stats *Stats) Median() float64 {
	return stats.Perc(50)
}

func Percentile(values []float64, perc float64) float64 {
	middle := float64(len(values)) * perc / 100.0
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
func (stats *Stats) Perc(perc float64) float64 {
	stats.Sort()
	return Percentile(stats.values, perc)
}

func (stats *Stats) String() string {
	return fmt.Sprintf("len: %d, avg: %.1f, med: %.1f, perc_95: %.1f, perc_99: %.1f, max: %.1f, min: %.1f",
		len(stats.values), stats.Avg(), stats.Median(), stats.Perc(95), stats.Perc(99),
		stats.Max(),
		stats.Min(),
	)
}

func (stats *Stats) Len() int {
	return len(stats.values)
}

func (stats *Stats) Variance() float64 {
	avg := stats.Avg()
	sum := 0.0
	for _, v := range stats.values {
		sum += math.Pow(v-avg, 2.0)
	}
	return sum / float64(stats.Len()-1)
}

func (s *Stats) StdDeviation() float64 {
	return math.Sqrt(s.Variance())
}

func (stats *Stats) Reset() {
	stats.values = stats.values[:0]
	stats.sorted = false
}
