package dockerclient

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var data = []struct {
	in  string
	out []string
}{
	{"foo", []string{"", "foo", ""}},
	{"foo/bar", []string{"foo", "bar", ""}},
	{"foo/bar:buz", []string{"foo", "bar", "buz"}},
	{"foo.boo/bar:buz", []string{"foo.boo", "bar", "buz"}},
	{"foo.boo:500/bar:buz", []string{"foo.boo:500", "bar", "buz"}},
}

func TestSplitImageName(t *testing.T) {
	for _, tt := range data {
		Convey(fmt.Sprintf("Given the image name %s", tt.in), t, func() {
			iname := tt.in
			Convey("When the name is split into segments", func() {
				registry, repository, tag := splitImageName(iname)
				Convey("Then the segments should be found properly", func() {
					So(registry, ShouldEqual, tt.out[0])
					So(repository, ShouldEqual, tt.out[1])
					So(tag, ShouldEqual, tt.out[2])
				})
			})
		})
	}
}
