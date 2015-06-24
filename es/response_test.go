package es

import (
	"encoding/json"
	"testing"

	"github.com/dynport/dgtk/es/aggregations"
)

func TestUnmarshalRequest(t *testing.T) {
	b := mustReadFixture(t, "response_with_aggregations.json")
	r := &Response{
		Aggregations: aggregations.Aggregations{},
	}
	failIfError(t, json.Unmarshal(b, r))
	failIfNil(t, r.Aggregations)
	assertEqual(t, len(r.Aggregations), 1)
	errorIfNil(t, r.Aggregations["days"])
}
