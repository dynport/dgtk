package dockerclient

import (
	"bytes"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseBuildResponse(t *testing.T) {
	Convey("Parse build response", t, func() {
		Convey("Parse json response", func() {
			r := bytes.NewBufferString(newResponse)
			streams := BuildResponse{}
			e := handleJSONStream(r, func(s *JSONMessage) {
				streams = append(streams, s)
			})
			So(e, ShouldBeNil)
			So(len(streams), ShouldEqual, 9)
			So(streams[0].Stream, ShouldEqual, "Step 1 : FROM ubuntu\n")
			So(streams.ImageId(), ShouldEqual, "0f101a4836f6")
		})
	})
}

const newResponse = `{"stream":"Step 1 : FROM ubuntu\n"}
{"stream":" ---\u003e 8dbd9e392a96\n"}
{"stream":"Step 2 : RUN apt-get update\n"}
{"stream":" ---\u003e Using cache\n"}
{"stream":" ---\u003e 30d9e1cb9bb8\n"}
{"stream":"Step 3 : RUN apt-get upgrade -y\n"}
{"stream":" ---\u003e Using cache\n"}
{"stream":" ---\u003e 0f101a4836f6\n"}
{"stream":"Successfully built 0f101a4836f6\n"}
`
