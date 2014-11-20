package main

import (
	"io/ioutil"
	"log"
	"testing"
)

var b = &build{Dir: "test-app"}

func init() {
	logger = log.New(ioutil.Discard, "", 0)
}

func TestDeps(t *testing.T) {
	deps, err := b.deps()
	if err != nil {
		t.Error("error getting deps", err)
	}

	if len(deps) != 23 {
		t.Errorf("expected to get 23 deps, got %d", len(deps))
	}
	d := deps[4]
	if d != "github.com/dynport/gocli" {
		t.Errorf("expected deps[4] to eq %q, was %q", "github.com/dynport/gocli", d)
	}
}

func TestCurrentPackage(t *testing.T) {
	cp, err := b.currentPackage()
	if err != nil {
		t.Fatal("error calling currentPackage", err)
	}
	if cp != "github.com/dynport/dgtk/go-remote-build/test-app" {
		t.Errorf("expected currentPackage to eq %q, was %q", "github.com/dynport/dgtk/go-remote-build/test-app", cp)
	}
}

func TestSplitBucket(t *testing.T) {
	tests := []struct {
		Bucket         string
		Key            string
		ExpectedBucket string
		ExpectedKey    string
	}{
		{"de-dynport-public/bin/linux_amd64/", "metrix", "de-dynport-public", "bin/linux_amd64/metrix"},
		{"de-dynport-public", "metrix", "de-dynport-public", "metrix"},
	}

	for _, tst := range tests {
		bucket, key := bucketAndKey(tst.Bucket, tst.Key)
		if bucket != tst.ExpectedBucket {
			t.Errorf("expected bucket to eq %q, was %q", tst.ExpectedBucket, bucket)
		}
		if key != tst.ExpectedKey {
			t.Errorf("expected key to eq %q, was %q", tst.ExpectedKey, key)
		}
	}
}

func TestFilesMap(t *testing.T) {
	_, err := b.filesMap()
	if err != nil {
		t.Fatal("error getting filesMap", err)
	}
}

func TestCreateArchive(t *testing.T) {
	_, err := b.createArchive()
	if err != nil {
		t.Error("error creating archive", err)
	}
}

func TestParseConfig(t *testing.T) {
	cfg, err := parseConfig("ubuntu@127.0.0.1")
	if err != nil {
		t.Fatal("error parsing config", err)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected Host to eq %q, was %q", "127.0.0.1", cfg.Host)
	}

	if cfg.User != "ubuntu" {
		t.Errorf("expected User to eq %q, was %q", "ububuntu", cfg.User)
	}

	cfg, err = parseConfig("ubuntu@1.2.3.4:1234")
	if err != nil {
		t.Fatal("error parsing config", err)
	}
	if cfg.Port != 1234 {
		t.Errorf("expected Port to eq 1234, was %d", cfg.Port)
	}

}
