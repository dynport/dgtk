package es

import (
	"io/ioutil"
	"testing"

	"github.com/dynport/dgtk/tskip/tskip"
)

func mustReadFixture(t *testing.T, name string) []byte {
	b, e := ioutil.ReadFile("fixtures/" + name)
	if e != nil {
		tskip.Errorf(t, 1, "unable to read fixute %q: %q", name, e)
		t.FailNow()
	}
	return b
}
