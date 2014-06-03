package github

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLink(t *testing.T) {
	Convey("ParseLink", t, func() {
		raw := `<https://api.github.com/user/26679/starred?page=2>; rel="next", <https://api.github.com/user/26679/starred?page=46>; rel="last"`
		link := ParseLink(raw)
		So(link, ShouldNotBeNil)
		So(link.Next, ShouldEqual, "https://api.github.com/user/26679/starred?page=2")
		So(link.Last, ShouldEqual, "https://api.github.com/user/26679/starred?page=46")

		Convey("more complex", func() {
			raw := `<https://api.github.com/user/26679/starred?page=3>; rel="next", <https://api.github.com/user/26679/starred?page=46>; rel="last", <https://api.github.com/user/26679/starred?page=1>; rel="first", <https://api.github.com/user/26679/starred?page=1>; rel="prev"`
			link := ParseLink(raw)
			So(link, ShouldNotBeNil)
			So(link.First, ShouldEqual, "https://api.github.com/user/26679/starred?page=1")
			So(link.Prev, ShouldEqual, "https://api.github.com/user/26679/starred?page=1")
			So(link.Next, ShouldEqual, "https://api.github.com/user/26679/starred?page=3")
			So(link.Last, ShouldEqual, "https://api.github.com/user/26679/starred?page=46")
		})
	})
}
