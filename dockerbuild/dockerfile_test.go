package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDockerfile(t *testing.T) {
	Convey("Dockerfile", t, func() {
		So(1, ShouldEqual, 1)
		file := NewDockerfile(file)
		So(file, ShouldNotBeNil)
		So(string(file), ShouldEqual, "FROM ubuntu\nRUN apt-get update\n")
		newFile := file.MixinProxy("http://127.0.0.1")
		So(string(newFile), ShouldEqual, "FROM ubuntu\nENV http_proxy http://127.0.0.1\nRUN apt-get update\n")
	})
}

var file = []byte(`FROM ubuntu
RUN apt-get update
`)
