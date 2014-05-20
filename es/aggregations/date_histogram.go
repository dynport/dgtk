package aggregations

import "encoding/json"

type DateHistogram struct {
	Name         string         `json:"name"`
	Field        string         `json:"field"`
	Interval     string         `json:"interval"`
	Aggregations json.Marshaler `json:"aggregations,omitempty"`
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
	return json.Marshal(hash{
		a.Name: h,
	},
	)
}
