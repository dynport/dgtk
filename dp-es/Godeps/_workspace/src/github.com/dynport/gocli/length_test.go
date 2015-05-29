package gocli

import "testing"

func TestLength(t *testing.T) {
	tab := NewTable()
	tab.Add("a", "ab")

	tests := []struct {
		Name     string
		Expected interface{}
		Value    interface{}
	}{
		{"0", 1, tab.Lengths[0]},
		{"1", 2, tab.Lengths[1]},
	}

	for _, tst := range tests {
		if tst.Expected != tst.Value {
			t.Errorf("expected %s to be %#v, was %#v", tst.Name, tst.Expected, tst.Value)
		}
	}

	tab.Add(Red("abc"), Green("abcd"))

	tests = []struct {
		Name     string
		Expected interface{}
		Value    interface{}
	}{
		{"0", 3, tab.Lengths[0]},
		{"1", 4, tab.Lengths[1]},
	}

	for _, tst := range tests {
		if tst.Expected != tst.Value {
			t.Errorf("expected %s to be %#v, was %#v", tst.Name, tst.Expected, tst.Value)
		}
	}
}
