package util

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func cleanup() {
	os.RemoveAll("tmp")
}

type testSource struct {
	reader io.Reader
}

func (t *testSource) Open() (io.Reader, error) {
	return t.reader, nil
}

func (t *testSource) Size() (int64, error) {
	return 0, nil
}

func TestFileCache(t *testing.T) {
	Convey("Open", t, func() {
		defer cleanup()
		br := strings.NewReader("just a string")
		r := &FileCache{Source: &testSource{reader: br}, Key: "some.key", Root: "tmp/cache"}

		b, e := ioutil.ReadAll(r)
		So(e, ShouldBeNil)
		So(string(b), ShouldEqual, "just a string")
		_, e = os.Stat("tmp/cache/some.key")
		So(e, ShouldBeNil)

		// it should store uncompressed
		b, e = ioutil.ReadFile("tmp/cache/some.key")
		So(e, ShouldBeNil)
		So(string(b), ShouldEqual, "just a string")
	})

	Convey("store compressed", t, func() {
		defer cleanup()
		buf := &bytes.Buffer{}
		e := func() error {
			gz := gzip.NewWriter(buf)
			defer gz.Close()
			_, e := fmt.Fprintf(gz, "just a compressed test")
			return e
		}()
		So(e, ShouldBeNil)

		f := &FileCache{Source: &testSource{reader: buf}, Key: "some.key", Root: "tmp/cache", SrcCompressed: true}
		b, e := ioutil.ReadAll(f)
		So(e, ShouldBeNil)
		So(string(b), ShouldEqual, "just a compressed test")
	})
}
