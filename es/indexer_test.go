package es

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func waitFor(checkEvery time.Duration, maxWait time.Duration, check func() bool) bool {
	checkTimer := time.NewTimer(checkEvery)
	maxWaitTimer := time.NewTimer(maxWait)
	for {
		select {
		case <-maxWaitTimer.C:
			return false
		case <-checkTimer.C:
			if check() {
				return true
			}
			checkTimer.Reset(checkEvery)
		}
	}
}

func TestIndexer(t *testing.T) {
	Convey("Indexer", t, func() {
		index.DeleteIndex()
		index.Refresh()
		indexer := &Indexer{Index: index, IndexEvery: 100 * time.Millisecond, BatchSize: 4}
		So(indexer, ShouldNotBeNil)
		ch := indexer.Start()
		ch <- &Doc{Source: Source{"Raw": "Line 1"}}
		ch <- &Doc{Source: Source{"Raw": "Line 2"}}
		ch <- &Doc{Source: Source{"Raw": "Line 3"}}
		index.Refresh()

		rsp, e := index.Search(nil)
		So(e, ShouldBeNil)
		So(rsp, ShouldNotBeNil)
		So(rsp.Hits.Total, ShouldEqual, 0)
		So(indexer.Stats.Runs, ShouldEqual, 0)

		check := waitFor(10*time.Millisecond, 500*time.Millisecond, func() bool {
			return (indexer.Stats.Runs == 1) && (indexer.Stats.IndexedDocs == 3)
		})
		So(check, ShouldBeTrue)
		if !check {
			t.Fatal("timeout waiting for indexing")
		}
		index.Refresh()
		So(indexer.Stats.Runs, ShouldEqual, 1)
		So(indexer.Stats.IndexedDocs, ShouldEqual, 3)
		rsp, e = index.Search(nil)
		So(e, ShouldBeNil)
		So(rsp, ShouldNotBeNil)
		So(rsp.Hits.Total, ShouldEqual, 3)
		ch <- &Doc{Source: Source{"Raw": "Line 4"}}
		ch <- &Doc{Source: Source{"Raw": "Line 5"}}
		ch <- &Doc{Source: Source{"Raw": "Line 6"}}
		ch <- &Doc{Source: Source{"Raw": "Line 7"}}

		check = waitFor(10*time.Millisecond, 1*time.Second, func() bool {
			return (indexer.Stats.Runs == 2) && (indexer.Stats.IndexedDocs == 7)
		})
		So(check, ShouldBeTrue)
		if !check {
			t.Fatal("timeout waiting for indexing")
		}
		index.Refresh()
		So(indexer.Stats.Runs, ShouldEqual, 2)
		So(indexer.Stats.IndexedDocs, ShouldEqual, 7)
		ch <- &Doc{Source: Source{"Raw": "Line 8"}}
		close(ch)

		check = waitFor(10*time.Millisecond, 1*time.Second, func() bool {
			return (indexer.Stats.Runs == 3) && (indexer.Stats.IndexedDocs == 8)
		})
		So(check, ShouldBeTrue)
		if !check {
			t.Fatal("timeout waiting for indexing")
		}
		index.Refresh()
		So(indexer.Stats.Runs, ShouldEqual, 3)
		So(indexer.Stats.IndexedDocs, ShouldEqual, 8)
	})
}
