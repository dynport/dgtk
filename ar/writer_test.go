package ar

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestArWriter(t *testing.T) {
	Convey("ArWriter", t, func() {
		So(1, ShouldEqual, 1)
		buf := &bytes.Buffer{}
		f := NewWriter(buf)
		f.WriteHeader(&Header{Name: "a.txt"})
		f.Write([]byte("a"))
		f.WriteHeader(&Header{Name: "b.txt"})
		f.Write([]byte("ab"))
		f.WriteHeader(&Header{Name: "c.txt"})
		f.Write([]byte("c"))
		s := buf.String()
		So(s, ShouldEndWith, "c")
		So(s, ShouldContainSubstring, "a\nb.txt/")
		So(s, ShouldContainSubstring, "abc.txt/")
	})
}
