package stats

import "testing"

func TestSort(t *testing.T) {
	a := Map{
		"a": &Value{Key: "a", Value: 1},
		"b": &Value{Key: "b", Value: 1},
		"c": &Value{Key: "c", Value: 2},
	}
	list := a.ReversedValues()
	tests := []struct{ Has, Want interface{} }{
		{len(list), 3},
		{list[0].Key, "c"},
		{list[1].Key, "a"},
		{list[2].Key, "b"},
	}
	for i, tc := range tests {
		if tc.Has != tc.Want {
			t.Errorf("%d: want=%#v has=%#v", i+1, tc.Want, tc.Has)
		}
	}
}
