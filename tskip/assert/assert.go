package assert

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dynport/dgtk/tskip/tskip"
)

func NewWithSkip(t *testing.T, skip int) *Assert {
	return &Assert{t: t, skip: skip}
}

func New(t *testing.T) *Assert {
	return NewWithSkip(t, 0)
}

type Assert struct {
	t    *testing.T
	skip int
}

func (s *Assert) FailIf(check bool, messages ...interface{}) {
	if check {
		if len(messages) == 0 {
			messages = append(messages, "failed!")
		}
		o := []string{}
		for _, m := range messages {
			o = append(o, fmt.Sprint(m))
		}
		tskip.Error(s.t, s.skip+1, strings.Join(o, " "))
		s.t.FailNow()
	}
}

func (s *Assert) Equal(a, b interface{}, messages ...interface{}) {
	if !isEqual(a, b) {
		tskip.Errorf(s.t, s.skip+2, diffMessage(messageString("not equal", messages), a, b))
	}
}

func (s *Assert) FailIfNil(i interface{}, messages ...interface{}) {
	if i == nil {
		tskip.Errorf(s.t, s.skip+1, messageString("expected not nil, was nil", messages))
		s.t.FailNow()
	}
}

func (s *Assert) ErrorIfNil(i interface{}, messages ...interface{}) {
	if i == nil {
		tskip.Errorf(s.t, s.skip+1, messageString("must be not nil, was nil", messages))
	}
}

func (s *Assert) FailIfError(err error) {
	if err != nil {
		tskip.Errorf(s.t, s.skip+1, "expected no error, got %s", err)
		s.t.FailNow()
	}
}

func (s *Assert) NoError(err error) {
	if err != nil {
		tskip.Errorf(s.t, s.skip+1, "expected no error, got %s", err)
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

func diffMessage(m string, a, b interface{}) string {
	return m + fmt.Sprintf("\nexpected (%T)\n   %#v\nactual   (%T)\n   %#v", b, b, a, a)
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
