package progress

import (
	"io/ioutil"
	"log"
	"strings"
	"testing"
)

func TestProgressString(t *testing.T) {
	l := log.New(ioutil.Discard, "", 0)
	p := Start(l, WithTotal(100))
	defer p.Close()
	p.IncBy(10)

	s := p.String()

	for _, b := range []string{"010/100"} {
		if !strings.Contains(s, b) {
			t.Errorf("expected string %q to contain %q", s, b)
		}
	}

	p = Start(l, WithTotal(99))
	defer p.Close()
	p.IncBy(10)
	s = p.String()
	for _, b := range []string{"10/99"} {
		if !strings.Contains(s, b) {
			t.Errorf("expected string %q to contain %q", s, b)
		}
	}
}
