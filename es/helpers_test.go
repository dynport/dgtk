package es

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dynport/dgtk/tskip/tskip"
)

func failIf(t *testing.T, check bool, messages ...interface{}) {
	if check {
		if len(messages) == 0 {
			messages = append(messages, "failed!")
		}
		o := []string{}
		for _, m := range messages {
			o = append(o, fmt.Sprint(m))
		}
		tskip.Error(t, 1, strings.Join(o, " "))
		t.FailNow()
	}
}

func waitFor(checkEvery time.Duration, maxWait time.Duration, check func() bool) bool {
	checkTimer := time.NewTimer(checkEvery)
	maxWaitTimer := time.NewTimer(maxWait)
	for {
		select {
		case <-maxWaitTimer.C:
			return false
		case <-checkTimer.C:
			if check() {
				return true
			}
			checkTimer.Reset(checkEvery)
		}
	}
}

// this is the implementation of assertEqual
func assertEqual(t *testing.T, a, b interface{}, messages ...interface{}) {
	if !isEqual(a, b) {
		tskip.Errorf(t, 1, diffMessage(messageString("not equal", messages), a, b))
	}
}

func failIfNil(t *testing.T, i interface{}, messages ...interface{}) {
	if t == nil {
		tskip.Errorf(t, 1, messageString("expected not nil, was nil", messages))
		t.FailNow()
	}
}

func errorIfNil(t *testing.T, i interface{}, messages ...interface{}) {
	if t == nil {
		tskip.Errorf(t, 1, messageString("must be not nil, was nil", messages))
	}
}

func isEqual(a, b interface{}) bool {
	if va, ok := castInt64(a); ok {
		if vb, ok := castInt64(b); ok {
			return va == vb
		}
	}
	return a == b
}

func castInt64(i interface{}) (int64, bool) {
	switch c := i.(type) {
	case int64:
		return c, true
	case int:
		return int64(c), true
	}
	return 0, false
}

func failIfError(t *testing.T, err error) {
	if err != nil {
		tskip.Errorf(t, 1, "expected no error, got %s", err)
		t.FailNow()
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		tskip.Errorf(t, 1, "expected no error, got %s", err)
	}
}

func diffMessage(m string, a, b interface{}) string {
	return m + fmt.Sprintf("\nexpected (%T)\n   %#v\nactual   (%T)\n   %#v", a, a, b, b)
}

// "not equal"
func messageString(defaultMessage string, messages []interface{}) string {
	if len(messages) == 0 {
		return defaultMessage
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
