package es

import (
	. "github.com/smartystreets/goconvey/convey"
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
	Convey("Test unmarshal field", t, func() {
		line := &Line{}
		So(source.Unmarshal(line), ShouldBeNil)
		So(line.Tag, ShouldEqual, "tag")
		So(line.Float, ShouldEqual, 0.1)
		So(line.Int, ShouldEqual, 10)
	})
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
	Convey("Unmarshal statistics", t, func() {
		sf := &StatisticalFacet{}
		So(sf.Load(stat), ShouldBeNil)
		So(sf.Type, ShouldEqual, "statistical")
		So(sf.Count, ShouldEqual, 1909)
		So(sf.Total, ShouldEqual, 7295467964)
		So(sf.Min, ShouldEqual, 110000.0)
		So(sf.Max, ShouldEqual, 15555000.0)
		So(sf.Mean, ShouldEqual, 3821617.5819800943)
		So(sf.SumOfSquares, ShouldEqual, 3.898752534119804e+16)
		So(sf.Variance, ShouldEqual, 5818248664852.342)
		So(sf.StdDeviation, ShouldEqual, 2412104.613165097)
	})
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
	Convey("Unmarshal Terms Facet", t, func() {
		f := &TermsFacet{}
		So(f.Load(stat), ShouldBeNil)
		So(f.Type, ShouldEqual, "terms")
		So(f.Total, ShouldEqual, 1909)
		So(f.Missing, ShouldEqual, 1)
		So(f.Other, ShouldEqual, 2)
		So(len(f.Terms), ShouldEqual, 3)
		term := f.Terms[0]
		So(term.Count, ShouldEqual, 1085)
		So(term.Term, ShouldEqual, "api/v1/photos#create")
	})
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
	Convey("Unmarshal Date histogram facet", t, func() {
		So(h.Load(stat), ShouldBeNil)
		So(h.Type, ShouldEqual, "date_histogram")
		So(len(h.Entries), ShouldEqual, 2)
		entry := h.Entries[0]
		So(entry.Time, ShouldEqual, 1384005600000)
		So(entry.Count, ShouldEqual, 43)
		So(entry.Min, ShouldEqual, 632000)
		So(entry.Max, ShouldEqual, 6.908e+06)
		So(entry.Total, ShouldEqual, 117121999)
		So(entry.TotalCount, ShouldEqual, 44)
		So(entry.Mean, ShouldEqual, 2723767.418604651)
	})
}
