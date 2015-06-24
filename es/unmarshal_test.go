package es

import (
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
	failIfError(t, source.Unmarshal(line))
	assertEqual(t, line.Tag, "tag")
	assertEqual(t, line.Float, 0.1)
	assertEqual(t, line.Int, 10)
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
	failIfError(t, sf.Load(stat))
	assertEqual(t, sf.Type, "statistical")
	assertEqual(t, sf.Count, 1909)
	assertEqual(t, sf.Total, 7295467964)
	assertEqual(t, sf.Min, 110000.0)
	assertEqual(t, sf.Max, 15555000.0)
	assertEqual(t, sf.Mean, 3821617.5819800943)
	assertEqual(t, sf.SumOfSquares, 3.898752534119804e+16)
	assertEqual(t, sf.Variance, 5818248664852.342)
	assertEqual(t, sf.StdDeviation, 2412104.613165097)
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
	failIfError(t, f.Load(stat))
	assertEqual(t, f.Type, "terms")
	assertEqual(t, f.Total, 1909)
	assertEqual(t, f.Missing, 1)
	assertEqual(t, f.Other, 2)
	assertEqual(t, len(f.Terms), 3)

	term := f.Terms[0]
	assertEqual(t, term.Count, 1085)
	assertEqual(t, term.Term, "api/v1/photos#create")
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
	failIfError(t, h.Load(stat))
	assertEqual(t, h.Type, "date_histogram")
	assertEqual(t, len(h.Entries), 2)
	entry := h.Entries[0]

	assertEqual(t, entry.Time, 1384005600000)
	assertEqual(t, entry.Count, 43)
	assertEqual(t, entry.Min, 632000.0)
	assertEqual(t, entry.Max, 6.908e+06)
	assertEqual(t, entry.Total, 117121999)
	assertEqual(t, entry.TotalCount, 44)
	assertEqual(t, entry.Mean, 2723767.418604651)
}
