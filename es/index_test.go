package es

import (
	"testing"
)

func TestCreateIndex(t *testing.T) {
	_ = index.DeleteIndex()
	_, err := index.CreateIndex(KeywordIndex())
	failIfError(t, err)

	index.PostObject(&TestLog{Tag: "unicorn", Host: "he-host1", Raw: "that is a test"})
	index.PostObject(&TestLog{Tag: "unicorn", Host: "he-host1"})
	index.PostObject(&TestLog{Tag: "unicorn", Host: "he-host2", Raw: "this is a line"})
	failIfError(t, index.Refresh())

	req := &Request{
		Facets: Facets{
			"hosts": &Facet{
				Terms: &FacetTerms{
					Field: "Host",
				},
			},
		},
	}
	res, err := index.Search(req)
	failIfError(t, err)
	facet := res.Facets["hosts"]
	if facet == nil {
		t.Fatal("hosts facet must not be nil")
	}
	assertEqual(t, facet.Total, 3)
	assertEqual(t, len(facet.Terms), 2)
	stats := map[interface{}]int{}
	for _, v := range facet.Terms {
		stats[v.Term] += v.Count
	}
	assertEqual(t, stats["he-host2"], 1)
	assertEqual(t, stats["he-host1"], 2)

	queryHost := map[string]string{
		"that": "he-host1",
		"this": "he-host2",
	}

	for query, host := range queryHost {
		req = &Request{Size: 10}
		q := &Query{}
		q.QueryString = &QueryString{Query: query}
		req.Query = q

		res, err = index.Search(req)
		failIfError(t, err)
		assertEqual(t, res.Hits.Total, 1)
		assertEqual(t, len(res.Hits.Hits), 1)
		assertEqual(t, res.Hits.Hits[0].Source["Host"], host)
	}
}
