package dockerbuild

import (
	"testing"
)

func TestDockerfile(t *testing.T) {
	file := NewDockerfile(file)
	var v, ex interface{}
	v = string(file)
	ex = "FROM ubuntu\nRUN apt-get update\n"
	if ex != v {
		t.Errorf("expected Dockerfile to be %#v, was %#v", ex, v)
	}
	newFile := file.MixinProxy("http://127.0.0.1")
	v = string(newFile)
	ex = "FROM ubuntu\nENV http_proxy http://127.0.0.1\nRUN apt-get update\n"
	if ex != v {
		t.Errorf("expected Dockerfile to be %#v, was %#v", ex, v)
	}
}

var file = []byte(`FROM ubuntu
RUN apt-get update
`)
