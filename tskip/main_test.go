package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/dynport/dgtk/tskip/tskip"
)

func TestFunc(t *testing.T) {
	if !doSimulate() {
		t.SkipNow()
	}
	tc := func(s, prefix, expected string) {
		if value := strings.TrimPrefix(s, prefix); value != expected {
			tskip.Errorf(t, 1, "TrimPrefix %q for %q should return %q, was %q", prefix, s, expected, value)
		}
	}
	tc("a test", "a", " test")
	tc("a test", "a test", "")
	tc("a test", "b", "")
}

func TestAssert(t *testing.T) {
	if !doSimulate() {
		t.SkipNow()
	}
	assertEqual(t, 1, 2, "number of records in the database")
	assertEqual(t, 1, 1)
	assertEqual(t, 1, int64(1))
	assertEqual(t, "test", "test")
}

// this is the implementation of assertEqual
func assertEqual(t *testing.T, a, b interface{}, messages ...interface{}) {
	if a != b {
		tskip.Errorf(t, 1, diffMessage(messageString(messages), a, b))
	}
}

func diffMessage(m string, a, b interface{}) string {
	return m + fmt.Sprintf("\nexpected (%T)\n   %#v\nactual   (%T)\n   %#v", a, a, b, b)
}

func messageString(messages []interface{}) string {
	if len(messages) == 0 {
		return "not equal"
	}
	o := []string{}
	for _, s := range messages {
		o = append(o, fmt.Sprint(s))
	}
	return strings.Join(o, " ")
}

func doSimulate() bool {
	return os.Getenv("TEST_SIMULATE") == "true"
}
