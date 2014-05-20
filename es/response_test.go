package es

import (
	"encoding/json"
	"testing"

	"github.com/dynport/dgtk/es/aggregations"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnmarshalRequest(t *testing.T) {
	Convey("Unmarshal Request", t, func() {
		b := mustReadFixture(t, "response_with_aggregations.json")
		So(b, ShouldNotBeNil)

		r := &Response{
			Aggregations: aggregations.Aggregations{},
		}
		So(json.Unmarshal(b, r), ShouldBeNil)
		So(r, ShouldNotBeNil)
		So(r.Aggregations, ShouldNotBeNil)
		So(len(r.Aggregations), ShouldEqual, 1)

		days := r.Aggregations["days"]
		So(days, ShouldNotBeNil)
	})
}
