package es

type Response struct {
	Took     int            `json:"took"`
	TimedOut bool           `json:"timed_out"`
	Facets   ResponseFacets `json:"facets"`
	Hits     Hits           `json:"hits"`
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
