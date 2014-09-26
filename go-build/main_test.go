package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuild(t *testing.T) {
	Convey("build", t, func() {
		p, e := build("test-proj")
		So(e, ShouldBeNil)
		So(p, ShouldNotBeNil)

		buf := &bytes.Buffer{}
		c := exec.Command(p)
		c.Stdout = buf
		c.Stderr = os.Stderr
		e = c.Run()
		So(e, ShouldBeNil)
		s := buf.String()
		So(s, ShouldContainSubstring, "Versions")

		status := &BuildStatus{}
		e = json.Unmarshal(buf.Bytes(), &status)
		So(e, ShouldBeNil)
		So(status.Name, ShouldEqual, "github.com/dynport/dgtk/go-build/test-proj")
	})
}
