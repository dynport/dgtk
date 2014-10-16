package main

import (
	"encoding/xml"
	"net/url"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var bucket = os.Getenv("TEST_BUCKET")

func TestS3(t *testing.T) {
	Convey("normalize params", t, func() {
		req, e := url.Parse("http://127.0.0.1/just/a")
		So(e, ShouldBeNil)
		normalized := normalizeParams(req)
		So(normalized, ShouldEqual, "")

		req, e = url.Parse("http://127.0.0.1/just/a?versionId=a&partNumber=1&acl=true")
		So(e, ShouldBeNil)
		So(req, ShouldNotBeNil)
		normalized = normalizeParams(req)
		So(normalized, ShouldEqual, "acl=true&partNumber=1&versionId=a")
	})

	Convey("upload", t, func() {
		if bucket == "" {
			t.Skip("no bucket defined")
		}
		client := &Client{
			Key:    os.Getenv("AWS_ACCESS_KEY_ID"),
			Secret: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		}
		s := "this is just a test"
		e := client.put(bucket, "test.txt", strings.NewReader(s), nil)
		So(e, ShouldBeNil)
	})

	Convey("upload stream", t, func() {
		if bucket == "" {
			t.Skip("no bucket defined")
		}
		f, e := os.Open("/Users/tobias/Downloads/Skype_6.11.60.455.dmg")
		So(e, ShouldBeNil)
		defer f.Close()
		e = client.put(bucket, "Skype_6.11.60.455.dmg", f, nil)
		So(e, ShouldBeNil)
	})

	Convey("Serialize Part", t, func() {
		p := &Part{PartNumber: 1, ETag: "test"}
		b, e := xml.Marshal(p)
		So(e, ShouldBeNil)
		So(string(b), ShouldEqual, "<Part><PartNumber>1</PartNumber><ETag>test</ETag></Part>")
	})
}
