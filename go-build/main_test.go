package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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
		So(len(status.Versions), ShouldNotEqual, 0)
		So(len(status.Dependencies), ShouldEqual, 1)
		So(status.Dependencies[0].Name, ShouldEqual, "github.com/dynport/gocli")
		So(len(status.Dependencies[0].Versions), ShouldNotEqual, 0)
	})

	Convey("gitChanges", t, func() {
		dirty := "test-proj/dirty.txt"
		os.RemoveAll(dirty)
		changes, e := gitChanges("test-proj")
		So(e, ShouldBeNil)
		So(changes, ShouldEqual, false)
		e = ioutil.WriteFile(dirty, []byte("dirty"), 0644)
		defer os.RemoveAll(dirty)
		So(e, ShouldBeNil)
		changes, e = gitChanges("test-proj")
		So(e, ShouldBeNil)
		So(changes, ShouldEqual, true)
	})

	Convey("SplitBucket", t, func() {
		bucket, key := splitBucket("some-bucket-name")
		So(bucket, ShouldEqual, "some-bucket-name")
		So(key, ShouldEqual, "")

		bucket, key = splitBucket("some-bucket-name/some/path")
		So(bucket, ShouldEqual, "some-bucket-name")
		So(key, ShouldEqual, "some/path")
	})
}
