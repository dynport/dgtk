package stats

import (
	"fmt"
	"testing"
)

func TestStdDeviation(t *testing.T) {
	s := New(20, 23, 23, 24, 25, 22, 12, 21, 29)

	f := func(in float64) string { return fmt.Sprintf("%.06f", in) }

	tests := []struct{ Has, Want interface{} }{
		{int(s.Max()), 29},
		{int(s.Min()), 12},
		{f(s.Sum()), "199.000000"},
		{f(s.Avg()), "22.111111"},
		{f(s.StdDeviation()), "4.594683"},
	}
	for i, tc := range tests {
		if tc.Want != tc.Has {
			t.Errorf("%d: want %#v, was %#v", i+1, tc.Want, tc.Has)
		}
	}
}
