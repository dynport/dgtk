package es

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

type TestLog struct {
	Id      int
	Tag     string
	Host    string
	Created string
	Raw     string
}

var sleepFor = 2 * time.Second

var index = &Index{
	Host:  "127.0.0.1",
	Port:  9200,
	Index: "test",
	Type:  "logs",
}

func validateEsRunning(t *testing.T) {
	_, e := index.Status()
	if e != nil {
		t.Fatalf("ElasticSearch is not running on %s. %s", index.BaseUrl(), e.Error())
	}
}

func setupIndex() error {
	index.DeleteIndex()
	index.EnqueueBulkIndex("1", &TestLog{Id: 1, Tag: "nginx", Created: "2013-12-03"})
	index.EnqueueBulkIndex("2", &TestLog{Id: 1, Tag: "unicorn", Created: "2013-12-03"})

	index.EnqueueBulkIndex("3", &TestLog{Id: 1, Tag: "nginx", Created: "2013-12-02"})
	index.EnqueueBulkIndex("4", &TestLog{Id: 1, Tag: "unicorn", Created: "2013-12-02"})

	index.EnqueueBulkIndex("5", &TestLog{Id: 1, Tag: "nginx", Created: "2013-12-01"})
	index.EnqueueBulkIndex("6", &TestLog{Id: 1, Tag: "unicorn", Created: "2013-12-01"})
	return index.RunBatchIndex()
}

func TestDeleteFromImage(t *testing.T) {
	validateEsRunning(t)
	Convey("Delete from Index", t, func() {
		So(setupIndex(), ShouldBeNil)
		So(index.Refresh(), ShouldBeNil)
		req := &Request{
			Size: 10,
		}
		rsp, e := index.Search(req)
		So(e, ShouldBeNil)
		So(rsp.Hits.Total, ShouldEqual, 6)

		_, r := index.DeleteByQuery("Tag:nginx")
		So(r, ShouldBeNil)
		So(index.Refresh(), ShouldBeNil)

		rsp, e = index.Search(req)
		So(e, ShouldBeNil)
		So(rsp.Hits.Total, ShouldEqual, 3)

		tags := map[string]int{}
		for _, hit := range rsp.Hits.Hits {
			tag := hit.Source["Tag"]
			switch tag := tag.(type) {
			case string:
				tags[tag]++
			}
		}
		So(tags["unicorn"], ShouldEqual, 3)

		_, r = index.DeleteByQuery("Tag:unicorn AND Created:[2013-12-01 TO 2013-12-02]")
		So(r, ShouldBeNil)
		So(index.Refresh(), ShouldBeNil)

		rsp, e = index.Search(req)
		So(e, ShouldBeNil)
		So(rsp.Hits.Total, ShouldEqual, 1)

		So(rsp.Hits.Hits[0].Source["Created"], ShouldEqual, "2013-12-03")
	})
}
