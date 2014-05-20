package aggregations

import "encoding/json"

type Terms struct {
	Name         string         `json:"name"`
	Field        string         `json:"field"`
	Aggregations json.Marshaler `json:"aggregations,omitempty"`
}

func (a *Terms) MarshalJSON() ([]byte, error) {
	h := hash{
		"terms": hash{
			"field": a.Field,
		},
	}
	if a.Aggregations != nil {
		h["aggregations"] = a
	}

	return json.Marshal(hash{
		a.Name: h,
	},
	)
}
