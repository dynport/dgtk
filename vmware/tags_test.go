package vmware

import (
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTags(t *testing.T) {
	var e error
	Convey("Tags", t, func() {
		tags := Tags{}
		So(tags.Len(), ShouldEqual, 0)

		tag := &Tag{VmId: "vm1", Key: "Name", Value: "This is the Name"}
		tags, e = tags.Update(tag)
		So(e, ShouldBeNil)
		So(tags.Len(), ShouldEqual, 1)
		So(tags[0].Key, ShouldEqual, "Name")
		So(tags[0].Value, ShouldEqual, "This is the Name")

		// updating an existing tag
		tag = &Tag{VmId: "vm1", Key: "Name", Value: "New Name"}
		tags, e = tags.Update(tag)
		So(e, ShouldBeNil)

		So(tags.Len(), ShouldEqual, 1)

		So(tags[0].Key, ShouldEqual, "Name")
		So(tags[0].Value, ShouldEqual, "New Name")

		// adding a new tag to the list
		tag = &Tag{VmId: "vm1", Key: "Enabled", Value: "true"}
		tags, e = tags.Update(tag)
		So(e, ShouldBeNil)
		So(tags.Len(), ShouldEqual, 2)

		sort.Sort(tags)

		So(tags[0].Key, ShouldEqual, "Enabled")
		So(tags[0].Value, ShouldEqual, "true")

		// removing a tag
		tag = &Tag{VmId: "vm1", Key: "Enabled", Value: ""}
		tags, e = tags.Update(tag)
		So(e, ShouldBeNil)
		So(tags.Len(), ShouldEqual, 1)
	})
}
