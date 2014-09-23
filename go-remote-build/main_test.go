package main

import (
	"io/ioutil"
	"log"
	"os"
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuild(t *testing.T) {
	b := &build{Dir: "test-app"}
	logger = log.New(ioutil.Discard, "", 0)

	Convey("deps", t, func() {
		deps, e := b.deps()
		So(e, ShouldBeNil)
		So(len(deps), ShouldBeGreaterThan, 10)
		sort.Strings(deps)
		So(deps, ShouldContain, "github.com/dynport/gocli")
	})

	Convey("currentPackage", t, func() {
		cp, e := b.currentPackage()
		So(e, ShouldBeNil)
		So(cp, ShouldEqual, "github.com/dynport/dgtk/go-remote-build/test-app")
	})

	Convey("filesMap", t, func() {
		m, e := b.filesMap()
		So(e, ShouldBeNil)
		So(len(m), ShouldBeGreaterThan, 3)
	})

	Convey("createarchive", t, func() {
		f, e := os.Create("/tmp/archive")
		So(e, ShouldBeNil)
		defer f.Close()

		file, e := b.createArchive()
		So(e, ShouldBeNil)
		So(file, ShouldNotEqual, "")
		defer os.RemoveAll(file)
	})
}
