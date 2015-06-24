package es

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

func TestParseStats(t *testing.T) {
	b := mustReadFixture(t, "stats.json")

	stats := &Stats{}
	assertNoError(t, json.Unmarshal(b, stats))
	names := stats.IndexNames()
	sort.Strings(names)

	assertEqual(t, strings.Join(names, ","), "logs,test")
	assertEqual(t, stats.Shards.Total, 20)
	assertEqual(t, stats.Shards.Successful, 20)
	assertEqual(t, stats.Shards.Failed, 0)

	assertEqual(t, len(stats.Indices), 2)
	index := stats.Indices["logs"]
	failIfNil(t, index, "logs index must not be nil")

	total := index.Total
	failIfNil(t, total)

	store := total.Store
	failIfNil(t, store)

	docs := total.Docs
	failIfNil(t, docs)

	assertEqual(t, docs.Count, 306672)
	assertEqual(t, docs.Deleted, 33430)
	assertEqual(t, store.SizeInBytes, 54063739)
	assertEqual(t, store.ThrottleTimeInMillis, 0)
}
