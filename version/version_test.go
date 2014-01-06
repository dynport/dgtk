package version

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestParseVersion(t *testing.T) {
	Convey("ParseVersion", t, func() {
		Convey("Parsing a version", func() {
			v := &Version{}
			So(v.Parse("1.2.3"), ShouldBeNil)
			So(v.Major, ShouldEqual, 1)
			So(v.Minor, ShouldEqual, 2)
			So(v.Patch, ShouldEqual, 3)
		})

		a, e := Parse("0.1.2")
		So(e, ShouldBeNil)
		Convey("comparing version "+a.String(), func() {
			versions := []string{"0.0.0", "0.0.9", "0.1.1"}
			for i, _ := range versions {
				v := versions[i]
				Convey(v+" should be less", func() {
					b, e := Parse(v)
					So(e, ShouldBeNil)
					So(a.Less(b), ShouldBeFalse)
				})
			}

			versions = []string{"1.0.0", "0.2.0", "0.1.3"}
			for i := range versions {
				v := versions[i]
				Convey(v+" should not be less", func() {
					b, e := Parse(v)
					So(e, ShouldBeNil)
					So(a.Less(b), ShouldBeTrue)
				})
			}
		})

		Convey("Compare Equal Version", func() {
			b, e := Parse("0.1.2")
			So(e, ShouldBeNil)
			So(a.Less(b), ShouldBeFalse)
		})
	})
}
