package es

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCreateIndex(t *testing.T) {
	Convey("Create index", t, func() {
		index.DeleteIndex()
		So(1, ShouldEqual, 1)
		b, e := json.Marshal(KeywordIndex())
		So(e, ShouldBeNil)
		So(len(b), ShouldNotEqual, 0)

		index.DeleteIndex()
		rsp, e := index.CreateIndex(KeywordIndex())
		So(e, ShouldBeNil)
		So(rsp, ShouldNotBeNil)

		index.PostObject(&TestLog{Tag: "unicorn", Host: "he-host1", Raw: "that is a test"})
		index.PostObject(&TestLog{Tag: "unicorn", Host: "he-host1"})
		index.PostObject(&TestLog{Tag: "unicorn", Host: "he-host2", Raw: "this is a line"})
		So(index.Refresh(), ShouldBeNil)

		req := &Request{
			Facets: Facets{
				"hosts": &Facet{
					Terms: &FacetTerms{
						Field: "Host",
					},
				},
			},
		}
		res, e := index.Search(req)
		So(e, ShouldBeNil)
		facet := res.Facets["hosts"]
		So(facet, ShouldNotBeNil)
		So(facet.Total, ShouldEqual, 3)
		So(len(facet.Terms), ShouldEqual, 2)
		stats := map[interface{}]int{}
		for _, v := range facet.Terms {
			stats[v.Term] += v.Count
		}
		So(stats["he-host2"], ShouldEqual, 1)
		So(stats["he-host1"], ShouldEqual, 2)

		queryHost := map[string]string{
			"that": "he-host1",
			"this": "he-host2",
		}

		for query, host := range queryHost {
			req = &Request{Size: 10}
			req.Query = &Query{}
			req.Query.QueryString = &QueryString{Query: query}

			res, e = index.Search(req)
			So(e, ShouldBeNil)
			So(res.Hits.Total, ShouldEqual, 1)
			So(len(res.Hits.Hits), ShouldEqual, 1)
			So(res.Hits.Hits[0].Source["Host"], ShouldEqual, host)
		}

	})
}
