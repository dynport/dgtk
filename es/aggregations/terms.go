package aggregations

import "encoding/json"

type Terms struct {
	Field        string                    `json:"field"`
	Order        map[string]string         `json:"sort,omitempty"`
	Aggregations map[string]json.Marshaler `json:"aggs,omitempty"`
	Size         int                       `json:"size"`
}

func (a *Terms) MarshalJSON() ([]byte, error) {
	h := hash{
		"field": a.Field,
	}
	if a.Order != nil {
		h["order"] = a.Order
	}
	h["size"] = a.Size
	//if a.Size != 0 {
	//	h["size"] = strconv.Itoa(a.Size)
	//}
	out := hash{"terms": h}
	if a.Aggregations != nil {
		out["aggs"] = a.Aggregations
	}
	return json.Marshal(out)
}
