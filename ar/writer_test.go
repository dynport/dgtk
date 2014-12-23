package ar

import (
	"bytes"
	"strings"
	"testing"
)

func TestArWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewWriter(buf)
	f.WriteHeader(&Header{Name: "a.txt"})
	f.Write([]byte("a"))
	f.WriteHeader(&Header{Name: "b.txt"})
	f.Write([]byte("ab"))
	f.WriteHeader(&Header{Name: "c.txt"})
	f.Write([]byte("c"))
	s := buf.String()

	if !strings.HasSuffix(s, "c") {
		t.Errorf("expected %q to have suffix", s, "c")
	}
	for _, ss := range []string{"a\nb.txt/", "abc.txt/"} {
		if !strings.Contains(s, ss) {
			t.Errorf("expected string %q to contain %q", s, ss)
		}
	}
}
