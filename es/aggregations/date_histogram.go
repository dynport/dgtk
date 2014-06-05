package aggregations

import "encoding/json"

type DateHistogram struct {
	Field        string                    `json:"field"`
	Interval     string                    `json:"interval"`
	Aggregations map[string]json.Marshaler `json:"aggregations,omitempty"`
}

func (a *DateHistogram) MarshalJSON() ([]byte, error) {
	h := hash{
		"date_histogram": hash{
			"field":    a.Field,
			"interval": a.Interval,
		},
	}
	if a.Aggregations != nil {
		h["aggregations"] = a.Aggregations
	}
	return json.Marshal(h)
}
