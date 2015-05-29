package gocli

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var router *Router

type DummyWriter struct {
}

func (writer *DummyWriter) Write([]byte) (int, error) {
	return 0, nil
}

func init() {
	if router == nil {
		router = &Router{}
		router.Writer = &DummyWriter{}
		args := &Args{}
		args.RegisterString("-h", "host", false, "127.0.0.1", "host to use")
		args.RegisterString("-i", "image", true, "", "Image id")
		router.Register(
			"ssh",
			&Action{
				Description: "SSH Into",
				Usage:       "<search>",
			},
		)
		router.Register(
			"container/start",
			&Action{
				Description: "start a container",
				Args:        args,
				Usage:       "<container_id>",
			},
		)
		router.Register(
			"container/stop",
			&Action{
				Description: "stop a container",
				Usage:       "<container_id>",
			},
		)
	}
}

func TestRouter(t *testing.T) {
	assert.NotNil(t, router)
	usage := router.Usage()
	assert.Contains(t, usage, "ssh      \t     \t<search>")
	assert.Contains(t, usage, "container\tstart\t<container_id>\tstart a container")
	assert.Contains(t, usage, "container\tstop \t<container_id>\tstop a container")
	assert.Contains(t, usage, `-h DEFAULT: "127.0.0.1" host to use`)
	assert.Contains(t, usage, `-i REQUIRED             Image id`)
}

func TestSearchActions(t *testing.T) {
	router := NewRouter(map[string]*Action{
		"container/start": {},
		"container/stop":  {},
		"image/list":      {},
	},
	)
	assert.Equal(t, len(router.Actions), 3)
	assert.Equal(t, len(router.Search([]string{})), 3)
	assert.Equal(t, len(router.Search([]string{"con"})), 2)
}

func TestMatchKey(t *testing.T) {
	assert.True(t, router.matchKey([]string{"co", "sta"}, "container/start"))
	assert.True(t, router.matchKey([]string{"co"}, "container/start"))
	assert.False(t, router.matchKey([]string{"ont", "sta"}, "container/start"))
	assert.False(t, router.matchKey([]string{"co", "sto"}, "container/start"))
}

func TestRouterSearch(t *testing.T) {
	assert.Equal(t, len(router.Search([]string{"con", "sta"})), 1)
	assert.Equal(t, len(router.Search([]string{"con", "st"})), 2)
	assert.Equal(t, len(router.Search([]string{"on", "st"})), 0)
}

func TestHandle(t *testing.T) {
	res := []string{}

	router := NewRouter(map[string]*Action{
		"container/start": {
			Handler: func(*Args) error {
				res = append(res, "container.start")
				return nil
			},
		},
		"container/stop": {},
		"image/list":     {},
	},
	)
	router.Writer = &DummyWriter{}

	pattern := []string{"/some/bin", "co", "sta"}

	router.Handle(pattern)
	assert.Equal(t, res, []string{"container.start"})

	res = []string{}
	router.Handle([]string{"/some/path", "co", "sta", "1"})
	assert.Equal(t, res, []string{"container.start"})

	res = []string{}
	router.Handle([]string{"/some/path", "co", "st"})
	assert.Equal(t, res, []string{})
}
