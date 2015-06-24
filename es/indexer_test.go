package es

import (
	"testing"
	"time"
)

func TestIndexer(t *testing.T) {
	index.DeleteIndex()
	if _, err := index.CreateIndex(KeywordIndex()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	indexer := &Indexer{Index: index, IndexEvery: 1 * time.Second, BatchSize: 4}
	ch := indexer.Start()
	ch <- &Doc{Source: Source{"Raw": "Line 1"}}
	ch <- &Doc{Source: Source{"Raw": "Line 2"}}
	ch <- &Doc{Source: Source{"Raw": "Line 3"}}
	index.Refresh()

	rsp, err := index.Search(nil)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, rsp.Hits.Total, 0)
	if indexer.Stats == nil {
		t.Fatal("Stats must not be nil")
	}
	assertEqual(t, indexer.Stats.Runs, int64(0))

	check := waitFor(10*time.Millisecond, 2*time.Second, func() bool {
		return (indexer.Stats.Runs == 1) && (indexer.Stats.IndexedDocs == 3)
	})
	if !check {
		t.Fatal("timeout waiting for indexing")
	}
	assertNoError(t, index.Refresh())
	assertEqual(t, indexer.Stats.Runs, int64(1))
	assertEqual(t, indexer.Stats.IndexedDocs, int64(3))
	rsp, err = index.Search(nil)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, rsp.Hits.Total, 3)
	ch <- &Doc{Source: Source{"Raw": "Line 4"}}
	ch <- &Doc{Source: Source{"Raw": "Line 5"}}
	ch <- &Doc{Source: Source{"Raw": "Line 6"}}
	ch <- &Doc{Source: Source{"Raw": "Line 7"}}

	check = waitFor(10*time.Millisecond, 1*time.Second, func() bool {
		return (indexer.Stats.Runs == 2) && (indexer.Stats.IndexedDocs == 7)
	})
	if !check {
		t.Fatal("timeout waiting for indexing")
	}
	assertNoError(t, index.Refresh())
	assertEqual(t, indexer.Stats.Runs, int64(2))
	assertEqual(t, indexer.Stats.IndexedDocs, int64(7))
	ch <- &Doc{Source: Source{"Raw": "Line 8"}}
	close(ch)

	check = waitFor(10*time.Millisecond, 1*time.Second, func() bool {
		return (indexer.Stats.Runs == 3) && (indexer.Stats.IndexedDocs == 8)
	})
	if !check {
		t.Fatal("timeout waiting for indexing")
	}
	assertNoError(t, index.Refresh())
	assertEqual(t, indexer.Stats.Runs, int64(3))
	assertEqual(t, indexer.Stats.IndexedDocs, int64(8))
}
