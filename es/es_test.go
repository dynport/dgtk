package es

import (
	"testing"
	"time"

	"github.com/dynport/dgtk/tskip/tskip"
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
		tskip.Errorf(t, 1, "ElasticSearch is not running on %s. %s", index.BaseUrl(), e.Error())
		t.FailNow()
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
	failIfError(t, setupIndex())
	failIfError(t, index.Refresh())
	req := &Request{
		Size: 10,
	}
	rsp, err := index.Search(req)
	failIfError(t, err)
	assertEqual(t, rsp.Hits.Total, 6)

	_, err = index.DeleteByQuery("Tag:nginx")
	failIfError(t, err)
	failIfError(t, index.Refresh())

	rsp, err = index.Search(req)
	failIfError(t, err)
	assertEqual(t, rsp.Hits.Total, 3)

	tags := map[string]int{}
	for _, hit := range rsp.Hits.Hits {
		tag := hit.Source["Tag"]
		switch tag := tag.(type) {
		case string:
			tags[tag]++
		}
	}
	assertEqual(t, tags["unicorn"], 3)

	_, err = index.DeleteByQuery("Tag:unicorn AND Created:[2013-12-01 TO 2013-12-02]")
	failIfError(t, err)
	failIfError(t, index.Refresh())

	rsp, err = index.Search(req)
	failIfError(t, err)
	assertEqual(t, rsp.Hits.Total, 1)
	assertEqual(t, rsp.Hits.Hits[0].Source["Created"], "2013-12-03")
}
