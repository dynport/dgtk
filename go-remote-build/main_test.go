package main

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/dynport/dgtk/expect"
)

var b = &build{Dir: "test-app"}

func init() {
	logger = log.New(ioutil.Discard, "", 0)
}

func TestDeps(t *testing.T) {
	expect := expect.New(t)
	deps, e := b.deps()
	expect(e).ToBeNil()
	expect(deps).ToHaveLength(23)
	expect(deps[4]).ToEqual("github.com/dynport/gocli")
}

func TestCurrentPackage(t *testing.T) {
	expect := expect.New(t)
	cp, e := b.currentPackage()
	expect(e).ToBeNil()
	expect(cp).ToEqual("github.com/dynport/dgtk/go-remote-build/test-app")

}

func TestSplitBucket(t *testing.T) {
	expect := expect.New(t)
	s := "de-dynport-public/bin/linux_amd64/"
	bucket, key := bucketAndKey(s, "metrix")
	expect(bucket).ToEqual("de-dynport-public")
	expect(key).ToEqual("bin/linux_amd64/metrix")

	bucket, key = bucketAndKey("de-dynport-public", "metrix")
	expect(bucket).ToEqual("de-dynport-public")
	expect(key).ToEqual("metrix")
}

func TestFilesMap(t *testing.T) {
	expect := expect.New(t)
	_, e := b.filesMap()
	expect(e).ToBeNil()
	//expect(m).ToHaveLength(0)
}

func TestCreateArchive(t *testing.T) {
	expect := expect.New(t)
	_, e := b.createArchive()
	expect(e).ToBeNil()
}

func TestParseConfig(t *testing.T) {
	expect := expect.New(t)
	cfg, e := parseConfig("ubuntu@127.0.0.1")
	expect(e).ToBeNil()
	expect(cfg).ToNotBeNil()
	expect(cfg.Host).ToEqual("127.0.0.1")
	expect(cfg.User).ToEqual("ubuntu")

	cfg, e = parseConfig("ubuntu@1.2.3.4:1234")
	expect(e).ToBeNil()
	expect(cfg.Port).ToEqual(1234)

}
