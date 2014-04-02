package godiff

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDiff(t *testing.T) {
	Convey("Diff", t, func() {
		type tc struct {
			A    interface{}
			B    interface{}
			Null bool
			Diff string
		}

		type M map[interface{}]interface{}
		cases := []tc{
			{"a", "a", true, ""},
			{"a", "b", false, "a != b"},
			{1, 1, true, ""},
			{1, 2, false, "1 != 2"},
			{1.0, 1.1, false, "1 != 1.1"},

			{[]string{"a"}, []string{"a"}, true, ""},
			{[]string{"a"}, []string{"a", "b"}, false, "[a] != [a b]"},
			{[]string{"a", "b"}, []string{"a", "b"}, true, ""},
			{[]string{"a", "c"}, []string{"b", "c"}, false, "0: a != b"},

			{[]interface{}{"a"}, []interface{}{"a"}, true, ""},
			{[]interface{}{"1"}, []interface{}{"a"}, false, "0: 1 != a"},

			{M{"a": 1}, "b", false, "map[a:1] != b"},
			{M{"a": 1}, M{"a": 1}, true, ""},
			{M{"a": 1}, M{"a": 2}, false, "a -> 1 != 2"},
			{M{"a": 1, "b": 2}, M{"a": 1}, false, "2 (value:b) in a but not in b"},
			{M{"a": 1}, M{"a": 1, "b": 2}, false, "2 (value:b) in b but not in a"},

			{M{"a": M{"b": 1}}, M{"a": M{"b": 2}}, false, "a -> b -> 1 != 2"},
		}
		for _, c := range cases {
			diff := Diff(c.A, c.B)
			if c.Null {
				So(diff, ShouldBeNil)
			} else {
				So(diff.Diff, ShouldEqual, c.Diff)
			}
		}
	})
}
