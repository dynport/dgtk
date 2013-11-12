package es

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type Line struct {
	Tag   string
	Float float64
	Int   int
}

var source = &Source{
	"Tag":   "tag",
	"Float": 0.1,
	"Int":   10,
}

func TestUnmarshalField(t *testing.T) {
	line := &Line{}
	e := source.Unmarshal(line)
	assert.Nil(t, e)
	assert.Equal(t, line.Tag, "tag")
	assert.Equal(t, line.Float, 0.1)
	assert.Equal(t, line.Int, 10)
}

func TestUnmarshalStatistical(t *testing.T) {
	stat := map[string]interface{}{
		"_type":          "statistical",
		"count":          1909,
		"total":          7295467964,
		"min":            110000,
		"max":            15555000,
		"mean":           3821617.5819800943,
		"sum_of_squares": 38987525341198040,
		"variance":       5818248664852.342,
		"std_deviation":  2412104.613165097,
	}
	sf := &StatisticalFacet{}
	assert.Nil(t, sf.Load(stat))
	assert.Equal(t, sf.Type, "statistical", "type")
	assert.Equal(t, sf.Count, 1909, "count")
	assert.Equal(t, sf.Total, 7295467964)
	assert.Equal(t, sf.Min, 110000.0)
	assert.Equal(t, sf.Max, 15555000.0)
	assert.Equal(t, sf.Mean, 3821617.5819800943)
	assert.Equal(t, sf.SumOfSquares, 3.898752534119804e+16)
	assert.Equal(t, sf.Variance, 5818248664852.342)
	assert.Equal(t, sf.StdDeviation, 2412104.613165097)
}

func TestUnmarshalTermsFacet(t *testing.T) {
	stat := map[string]interface{}{
		"_type":   "terms",
		"missing": 1,
		"total":   1909,
		"other":   2,
		"terms": []map[string]interface{}{
			{
				"term":  "api/v1/photos#create",
				"count": 1085,
			},
			{
				"term":  "api/v1/my/photos#create",
				"count": 823,
			},
			{
				"term":  "photos#create",
				"count": 1,
			},
		},
	}
	f := &TermsFacet{}
	assert.Nil(t, f.Load(stat))
	assert.Equal(t, f.Type, "terms")
	assert.Equal(t, f.Total, 1909)
	assert.Equal(t, f.Missing, 1)
	assert.Equal(t, f.Other, 2)
	assert.Equal(t, len(f.Terms), 3)
	term := f.Terms[0]
	assert.Equal(t, term.Count, 1085)
	assert.Equal(t, term.Term, "api/v1/photos#create")
}

func TestUnmarshalDateHistogramFacet(t *testing.T) {
	stat := map[string]interface{}{
		"_type": "date_histogram",
		"entries": []map[string]interface{}{
			{
				"time":        1384005600000,
				"count":       43,
				"min":         632000,
				"max":         6908000,
				"total":       117121999,
				"total_count": 44,
				"mean":        2723767.418604651,
			},
			{
				"time":        1384006200000,
				"count":       40,
				"min":         668000,
				"max":         14600000,
				"total":       105379000,
				"total_count": 40,
				"mean":        2634475,
			},
		},
	}
	h := &DateHistogramFacet{}
	assert.Nil(t, h.Load(stat))
	assert.Equal(t, h.Type, "date_histogram")
	assert.Equal(t, len(h.Entries), 2)
	entry := h.Entries[0]
	assert.Equal(t, entry.Time, 1384005600000, "time")
	assert.Equal(t, entry.Count, 43, "count")
	assert.Equal(t, entry.Min, 632000, "min")
	assert.Equal(t, entry.Max, 6.908e+06, "max")
	assert.Equal(t, entry.Total, 117121999, "total")
	assert.Equal(t, entry.TotalCount, 44, "total_count")
	assert.Equal(t, entry.Mean, 2723767.418604651, "mean")
}
