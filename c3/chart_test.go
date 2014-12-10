package c3

import "testing"

func TestCharts(t *testing.T) {
	c := New()

	c.Push("a", 1)
	c.Push("a", 2)
	c.Push("b", 3)

	if len(c.Data.Columns) != 2 {
		t.Fatalf("expected to find 2 columns, found %d", len(c.Data.Columns))
	}

	if len(c.Data.Columns[0]) != 3 {
		t.Fatalf("expected to find 3 rows for 1 column, found %d", len(c.Data.Columns[0]))
	}

	for i, e := range []interface{}{"a", 1, 2} {
		v := c.Data.Columns[0][i]
		if v != e {
			t.Errorf("expected value at index %d to eq %q, was %q", i, e, v)
		}
	}

	for i, e := range []interface{}{"b", 3} {
		v := c.Data.Columns[1][i]
		if v != e {
			t.Errorf("expected value at index %d to eq %q, was %q", i, e, v)
		}
	}
}
