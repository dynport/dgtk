package es

type Entries []*Entry

type Entry struct {
	TimestampMicros int64 `json:"time"`
	Count           int   `json:"count"`
}

func (entry *Entry) Timestamp() int64 {
	return entry.TimestampMicros / 1000
}
