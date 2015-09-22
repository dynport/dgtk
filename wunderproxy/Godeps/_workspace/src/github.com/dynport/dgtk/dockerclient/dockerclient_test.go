package dockerclient

import (
	"testing"
)

func TestSplitImageName(t *testing.T) {
	tests := []struct {
		in         string
		registry   string
		repository string
		tag        string
	}{
		{"foo", "", "foo", ""},
		{"foo/bar", "foo", "bar", ""},
		{"foo/bar:buz", "foo", "bar", "buz"},
		{"foo.boo/bar:buz", "foo.boo", "bar", "buz"},
		{"foo.boo:500/bar:buz", "foo.boo:500", "bar", "buz"},
	}

	for _, tc := range tests {
		registry, repository, tag := splitImageName(tc.in)
		if registry != tc.registry {
			t.Errorf("expected registry to be %q, was %q", tc.registry, registry)
		}
		if repository != tc.repository {
			t.Errorf("expected repository to be %q, was %q", tc.repository, repository)
		}
		if tag != tc.tag {
			t.Errorf("expected tag to be %q, was %q", tc.tag, tag)
		}
	}
}
