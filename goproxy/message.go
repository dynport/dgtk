package goproxy

import (
	"time"
)

type Message struct {
	Type     string
	Status   string
	started  time.Time
	Duration time.Duration
	Resource *Resource
	Error    error
}

func NewMessage(t string, r *Resource) *Message {
	return &Message{Type: t, Resource: r, started: time.Now()}
}

type subscriber interface {
	Publish(*Message)
}

var subscribers = []subscriber{}

func (m Message) publishError(e error) {
	m.Error = e
	m.publish("error")
}

func (m Message) publish(status string) {
	m.Type += "." + status
	m.Duration = time.Since(m.started)
	for _, s := range subscribers {
		s.Publish(&m)
	}
}
