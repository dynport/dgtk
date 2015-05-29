package gocli

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntLength(t *testing.T) {
	tests := []struct {
		Integer  int
		Expected interface{}
	}{
		{1, 1},
		{0, 1},
		{10, 2},
		{100, 3},
		{-1, 2},
	}

	for _, tst := range tests {
		v := intLength(tst.Integer)
		if tst.Expected != v {
			t.Errorf("expected %d to be %#v, was %#v", tst.Integer, tst.Expected, v)
		}
	}
}

func TestStringLength(t *testing.T) {
	tests := []struct {
		Name     string
		Expected interface{}
	}{
		{"A", 1},
		{"Ü", 1},
		{"ÜÄÖ", 3},
		{Red("ÜÖ"), 2},
	}

	for _, tst := range tests {
		v := stringLength(tst.Name)
		if tst.Expected != v {
			t.Errorf("expected %s to be %#v, was %#v", tst.Name, tst.Expected, v)
		}
	}
}

func TestSortTable(t *testing.T) {
	ta := NewTable()
	ta.Add("a", 2)
	ta.Add("b", 1)

	ta.SortBy = 0
	sort.Sort(ta)

	assert.Equal(t, ta.Columns[0][0], "a")
	assert.Equal(t, ta.Columns[1][0], "b")

	ta.SortBy = 1
	sort.Sort(ta)
	assert.Equal(t, ta.Columns[0][0], "b")
	assert.Equal(t, ta.Columns[1][0], "a")
}

func TestTable(t *testing.T) {
	table := NewTable()
	assert.NotNil(t, table)
	table.Add("a", "b")
	table.Add("aa", "bb")
	assert.Equal(t, len(table.Lines(false)), 2)
	assert.Contains(t, table.String(), "a\tb")

	table.Separator = " "
	assert.Contains(t, table.String(), "a b")
}

func TestColorize(t *testing.T) {
	str := Colorize(90, "test")
	assert.NotNil(t, str)
}

func TestAddColumnsNotBeingStrings(t *testing.T) {
	table := NewTable()
	table.Separator = " "
	table.Add(1, 2, "a")
	assert.Contains(t, table.String(), "1 2 a")
}
