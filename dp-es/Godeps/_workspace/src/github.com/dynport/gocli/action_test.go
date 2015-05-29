package gocli

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParseAction(t *testing.T) {
	a := &Action{
	}
	t.Log(a)
	assert.NotNil(t, a)
}
