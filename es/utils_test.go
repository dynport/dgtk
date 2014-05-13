package es

import (
	"io/ioutil"
	"testing"
)

func mustReadFixture(t *testing.T, name string) []byte {
	b, e := ioutil.ReadFile("fixtures/" + name)
	if e != nil {
		t.Fatalf("unable to read fixute %q: %q", name, e)
	}
	return b
}
