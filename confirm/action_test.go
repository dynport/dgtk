package confirm

import "testing"

func TestAdd(t *testing.T) {
	list := Actions{}
	list.Create("create", nil, func() error { return nil })
	if v, ex := len(list), 1; ex != v {
		t.Errorf("expected len(list) to be %d, was %d", ex, v)
	}
}
