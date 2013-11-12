package es

import (
	"fmt"
)

type Facet struct {
	Type           string      `json:"_type,omitempty"`
	Missing        int         `json:"missing,omitempty"`
	Total          int         `json:"total,omitempty"`
	Other          int         `json:"other,omitempty"`
	Terms          *FacetTerms `json:"terms,omitempty"`
	*Entries       `json:"entries,omitempty"`
	*DateHistogram `json:"date_histogram,omitempty"`
	Statistical    *StatisticalFacet
}

type FacetTerms struct {
	Field   string        `json:"field,omitempty"`
	Size    int           `json:"size,omitempty"`
	Exclude []interface{} `json:"exclude,omitempty"`
}

type StatisticalFacet struct {
	Type         string  `json:"_type,omitempty"`
	Count        int64   `json:"count,omitempty"`
	Total        int64   `json:"total,omitempty"`
	Min          float64 `json:"min,omitempty"`
	Max          float64 `json:"max,omitempty"`
	Mean         float64 `json:"mean,omitempty"`
	SumOfSquares float64 `json:"sum_of_squares,omitempty"`
	Variance     float64 `json:"variance,omitempty"`
	StdDeviation float64 `json:"std_deviation,omitempty"`
}

func interfaceToInt64(i interface{}) int64 {
	switch v := i.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	default:
		panic(fmt.Sprintf("unable to convert %T %v to float64", v, v))
	}
}

func interfaceToFloat64(i interface{}) float64 {
	switch v := i.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	default:
		panic(fmt.Sprintf("unable to convert %T %v to float64", v, v))
	}
}

func (sf *StatisticalFacet) Load(m map[string]interface{}) error {
	sf.Type, _ = m["_type"].(string)
	sf.Count = interfaceToInt64(m["count"])
	sf.Total = interfaceToInt64(m["total"])
	sf.Min = interfaceToFloat64(m["min"])
	sf.Max = interfaceToFloat64(m["max"])
	sf.Mean = interfaceToFloat64(m["mean"])
	sf.SumOfSquares = interfaceToFloat64(m["sum_of_squares"])
	sf.Variance = interfaceToFloat64(m["variance"])
	sf.StdDeviation = interfaceToFloat64(m["std_deviation"])
	return nil
}

type TermsFacet struct {
	Type    string             `json:"_type,omitempty"`
	Missing int64              `json:"missing,omitempty"`
	Total   int64              `json:"total,omitempty"`
	Other   int64              `json:"other,omitempty"`
	Terms   FacetResponseTerms `json:"terms,omitempty"`
}

type FacetResponseTerms []*FacetResponseTerm

type FacetResponseTerm struct {
	Term  interface{}
	Count int64
}

func (tf *TermsFacet) Load(m map[string]interface{}) error {
	tf.Type, _ = m["_type"].(string)
	tf.Missing = interfaceToInt64(m["missing"])
	tf.Total = interfaceToInt64(m["total"])
	tf.Other = interfaceToInt64(m["other"])
	terms, ok := m["terms"].([]map[string]interface{})
	if ok {
		tf.Terms = make(FacetResponseTerms, 0, len(terms))
		for _, term := range terms {
			t := &FacetResponseTerm{
				Count: interfaceToInt64(term["count"]),
				Term:  term["term"],
			}
			tf.Terms = append(tf.Terms, t)
		}
	}
	return nil
}

type DateHistogramFacet struct {
	Type    string                     `json:"term,omitempty"`
	Entries []*DateHistogramFacetEntry `json:"entries,omitempty"`
}

type DateHistogramFacetEntry struct {
	Time       int64 `json:"time,omitempty"`
	Count      int64 `json:"time,omitempty"`
	Min        float64
	Max        float64
	Total      int64
	TotalCount int64
	Mean       float64
}

func (dhf *DateHistogramFacet) Load(m map[string]interface{}) error {
	dhf.Type, _ = m["_type"].(string)
	entries, ok := m["entries"].([]map[string]interface{})
	if ok {
		dhf.Entries = make([]*DateHistogramFacetEntry, 0, len(entries))
		for _, e := range entries {
			entry := &DateHistogramFacetEntry{
				Time:       interfaceToInt64(e["time"]),
				Count:      interfaceToInt64(e["count"]),
				Min:        interfaceToFloat64(e["min"]),
				Max:        interfaceToFloat64(e["max"]),
				Total:      interfaceToInt64(e["total"]),
				TotalCount: interfaceToInt64(e["total_count"]),
				Mean:       interfaceToFloat64(e["mean"]),
			}
			dhf.Entries = append(dhf.Entries, entry)
		}
	}
	return nil
}
