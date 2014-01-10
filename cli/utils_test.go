package cli

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTagMapValidation(t *testing.T) {
	Convey("Given a tag mapping", t, func() {
		tagM := map[string]string{"foo": "f", "bar": "b"}
		Convey(`When validation is called for allowed tags "foo", "bar"`, func() {
			result := validateTagMap(tagM, "foo", "bar")
			Convey("Then the result is nil", func() {
				So(result, ShouldBeNil)
			})
		})
		Convey(`When validation is called for allowed tags "foo", "bar", "baz" (last one unused)`, func() {
			result := validateTagMap(tagM, "foo", "bar", "baz")
			Convey("Then the result is nil", func() {
				So(result, ShouldBeNil)
			})
		})
		Convey(`When validation is called for allowed tag "unknown"`, func() {
			result := validateTagMap(tagM, "unknown")
			Convey("Then the result is not nil", func() {
				So(result, ShouldNotBeNil)
			})
			Convey("And the result should contain an error", func() {
				So(result.Error(), ShouldEqual, `unknown tag "foo"`)
			})
		})
	})
}
