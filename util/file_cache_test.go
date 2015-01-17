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
	defer cleanup()
	br := strings.NewReader("just a string")
	fc := &FileCache{Source: &testSource{reader: br}, Key: "some.key", Root: "tmp/cache"}

	r, err := fc.Open()
	if err != nil {
		t.Fatal("error opening file cache", err)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal("error reading file cache", err)
	}
	ex := "just a string"
	v := string(b)
	if v != ex {
		t.Errorf("expected %v to eq %v", v, ex)
	}
	var p string = "tmp/cache/some.key"
	_, err = os.Stat(p)
	if err != nil {
		t.Errorf("expected %v to exist", p)
	}

	// it should store uncompressed
	v = "tmp/cache/some.key"
	b, err = ioutil.ReadFile(v)
	if err != nil {
		t.Fatalf("error reading file %v: %v", v, err)

	}
	v = string(b)
	ex = "just a string"
	if ex != v {
		t.Errorf("expected string to be %#v, was %#v", ex, v)
	}
}

func TestReadCompressed(t *testing.T) {
	defer cleanup()
	buf := &bytes.Buffer{}
	err := func() error {
		gz := gzip.NewWriter(buf)
		defer gz.Flush()
		defer gz.Close()
		_, e := fmt.Fprintf(gz, "just a compressed test")
		return e
	}()
	if err != nil {
		t.Fatal("error writing to buffer", err)
	}
	fc := &FileCache{Source: &testSource{reader: buf}, Key: "some.key", Root: "tmp/cache", SrcCompressed: true}
	r, err := fc.Open()
	if err != nil {
		t.Fatalf("error opening file cache (len was %d): %v", buf.Len(), err)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal("error reading from file cache", err)
	}
	v := string(b)
	ex := "just a compressed test"
	if ex != v {
		t.Errorf("expected compressed string to be %#v, was %#v", ex, v)
	}
}
