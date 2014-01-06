package dockerclient

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestServerVersion(t *testing.T) {
	Convey("Version", t, func() {
		v := &Version{Version: "0.6.9"}
		So(v.jsonBuildStream(), ShouldBeFalse)

		v = &Version{Version: "0.7.0"}
		So(v.jsonBuildStream(), ShouldBeTrue)
	})
}
