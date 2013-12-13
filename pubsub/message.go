package pubsub

import (
	"time"
)

func NewMessage(key string, payload interface{}) *Message {
	return &Message{key: key, payload: payload, createdAt: time.Now()}
}

type Message struct {
	key       string
	createdAt time.Time
	payload   interface{}
}

func (message *Message) Payload() interface{} {
	return message.payload
}

func (message *Message) Key() string {
	return message.key
}

func (message *Message) CreatedAt() time.Time {
	return message.createdAt
}
