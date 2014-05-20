package es

import "github.com/dynport/dgtk/es/aggregations"

func NewResponse(raw []byte) *Response {
	return &Response{
		Raw:          raw,
		Aggregations: aggregations.Aggregations{},
	}
}

type Response struct {
	Took         int                       `json:"took"`
	TimedOut     bool                      `json:"timed_out"`
	Facets       ResponseFacets            `json:"facets"`
	Hits         Hits                      `json:"hits"`
	Aggregations aggregations.Aggregations `json:"aggregations"`

	Raw []byte `json:"-"`
}

type ResponseFacets map[string]*ResponseFacet

type ResponseFacet struct {
	Type    string       `json:"_type"`
	Missing int          `json:"missing"`
	Total   int          `json:"total"`
	Other   int          `json:"other"`
	Terms   []*FacetTerm `json:"terms,omitempty"`
	Entries []*Entry     `json:"entries,omitempty"`
}

type FacetTerm struct {
	Term  interface{} `json:"term"`
	Count int         `json:"count"`
}
