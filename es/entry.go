package es

type Entries []*Entry

type Entry struct {
	TimestampMicros int64   `json:"time"`
	Count           int64   `json:"count"`
	Min             float64 `json:"min,omitempty"`
	Max             float64 `json:"max,omitempty"`
	Total           float64 `json:"total,omitempty"`
	TotalCount      int64   `json:"total_count,omitempty"`
	Mean            float64 `json:"mean,omitempty"`
}

func (entry *Entry) Timestamp() int64 {
	return entry.TimestampMicros / 1000
}
