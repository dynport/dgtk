package dockerclient

import (
	"testing"
)

func TestHandleMessage(t *testing.T) {
	header := []byte{
		0, 0, 0, 0,
		0, 0, 31, 178,
	}
	var v, ex interface{} = messageLength(header), 8114
	if v != ex {
		t.Errorf("expected messageLength(header) to be %d, was %d", ex, v)
	}
}
