package dockerclient

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestParseBuildResponse(t *testing.T) {
	Convey("Parse build response", t, func() {
		Convey("Parse old response", func() {
			r := bytes.NewBufferString(outputPre0_7)
			dh := &DockerHost{}
			streams := BuildResponse{}
			res, e := dh.handleBuildImagePlain(r, func(s *Stream) {
				streams = append(streams, s)
			})
			So(e, ShouldBeNil)
			So(len(streams), ShouldEqual, 9)
			So(len(res), ShouldEqual, 9)
			So(streams[0].Stream, ShouldEqual, "Step 1 : FROM ubuntu\n")
			So(streams.ImageId(), ShouldEqual, "b30eb4fbfc51")
		})
		Convey("Parse json response", func() {
			r := bytes.NewBufferString(newResponse)
			dh := &DockerHost{}
			streams := BuildResponse{}
			res, e := dh.handleBuildImageJson(r, func(s *Stream) {
				streams = append(streams, s)
			})
			So(e, ShouldBeNil)
			So(len(streams), ShouldEqual, 9)
			So(len(res), ShouldEqual, 9)
			So(streams[0].Stream, ShouldEqual, "Step 1 : FROM ubuntu\n")
			So(streams.ImageId(), ShouldEqual, "0f101a4836f6")
		})
	})
}

const outputPre0_7 = `Step 1 : FROM ubuntu
---> 8dbd9e392a96
Step 2 : RUN apt-get update
---> Using cache
---> f7ada547a49b
Step 3 : RUN apt-get upgrade -y
---> Using cache
---> b30eb4fbfc51
Successfully built b30eb4fbfc51
`

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
