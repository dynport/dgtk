package es

type Stats struct {
	Shards  *Shards `json:"_shards,omitempty"`
	Indices map[string]*indexStats
}

func (stats *Stats) IndexNames() []string {
	names := make([]string, 0, len(stats.Indices))
	for name := range stats.Indices {
		names = append(names, name)
	}
	return names
}

type indexStats struct {
	Total *indexStatsDetail `json:"total,omitempty"`
}

type indexStatsDetail struct {
	Docs  *indexStatsDocs
	Store *indexStatsStore
}

type indexStatsDocs struct {
	Count   int64 `json:"count"`
	Deleted int64 `json:"deleted"`
}

type indexStatsStore struct {
	SizeInBytes          int64 `json:"size_in_bytes"`
	ThrottleTimeInMillis int64 `json:"throttle_time_in_millis"`
}
