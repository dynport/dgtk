package gosql

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMap(t *testing.T) {
	Convey("map", t, func() {
		So(1, ShouldEqual, 1)
		m := Map{"id": 1}
		So(m, ShouldNotBeNil)
		cols, e := m.Columns()
		So(e, ShouldBeNil)
		So(len(cols), ShouldEqual, 1)
		So(cols[0].Name, ShouldEqual, "id")

		names := cols.Names()
		So(len(names), ShouldEqual, 1)
		So(names[0], ShouldEqual, "id")

		values := cols.Values()
		So(len(values), ShouldEqual, 1)
		So(values[0], ShouldEqual, 1)

		m = Map{"id": 1, "name": "First Entry", "a": 1}

		cols, e = m.Columns("name")
		So(e, ShouldBeNil)
		So(len(cols), ShouldEqual, 1)

		vals := cols.Values()
		So(len(vals), ShouldEqual, 1)
		So(vals[0], ShouldEqual, "First Entry")

		all, e := m.Columns()
		So(e, ShouldBeNil)
		So(len(all), ShouldEqual, 3)
		So(all[0].Name, ShouldEqual, "a")
	})
}
