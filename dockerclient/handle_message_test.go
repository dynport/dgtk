package dockerclient

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestHandleMessage(t *testing.T) {
	Convey("Handle Message", t, func() {
		header := []byte{
			0, 0, 0, 0,
			0, 0, 31, 178,
		}
		So(messageLength(header), ShouldEqual, 8114)
	})
}
