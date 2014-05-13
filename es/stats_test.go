package es

import (
	"encoding/json"
	"strings"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestParseStats(t *testing.T) {
	Convey("Parse Stats", t, func() {
		So(1, ShouldEqual, 1)
		b := mustReadFixture(t, "stats.json")

		stats := &Stats{}
		e := json.Unmarshal(b, stats)
		So(e, ShouldBeNil)

		So(strings.Join(stats.IndexNames(), ","), ShouldEqual, "logs,test")

		So(stats, ShouldNotBeNil)
		So(stats.Shards.Total, ShouldEqual, 20)
		So(stats.Shards.Successful, ShouldEqual, 20)
		So(stats.Shards.Failed, ShouldEqual, 0)

		So(len(stats.Indices), ShouldEqual, 2)
		So(stats.Indices["logs"], ShouldNotBeNil)

		index := stats.Indices["logs"]
		if index == nil {
			t.Fatal("expected index logs to not be nil")
		}
		if index.Total == nil {
			t.Fatal("expected Total to not be nil")
		}
		total := index.Total
		if total.Docs == nil {
			t.Fatal("Docs is null")
		}
		docs := total.Docs
		So(docs.Count, ShouldEqual, 306672)
		So(docs.Deleted, ShouldEqual, 33430)

		store := total.Store
		if store == nil {
			t.Fatal("Store is null")
		}
		So(store.SizeInBytes, ShouldEqual, 54063739)
		So(store.ThrottleTimeInMillis, ShouldEqual, 0)

	})
}
