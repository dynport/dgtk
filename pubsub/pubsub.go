package pubsub

import (
	"fmt"
	"time"
)

type PubSub struct {
	subscribers []chan *Message
	Stats
}

func cloneMessage(from *Message) *Message {
	m := &Message{}
	if from != nil {
		m.CreatedAt = from.CreatedAt
		m.Duration = from.Duration
		m.Level = from.Level
		m.Payload = from.Payload
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return m
}

func (pubsub *PubSub) Publish(key, messageText string, message *Message) error {
	m := cloneMessage(message)
	m.Key = key
	m.Message = messageText
	pubsub.Stats.MessageReceived()
	var e error
	for _, s := range pubsub.subscribers {
		select {
		case s <- m:
			pubsub.Stats.MessageDispatched()
		default:
			e = fmt.Errorf("unable to publish to %v", s)
		}
	}
	return e
}

func (pubsub *PubSub) SubscribersCount() int {
	return len(pubsub.subscribers)
}

//  could not be bother to fight with locking just now
// func (pubsub *PubSub) Unsubscribe(c chan *Message) {
// implement me
// }

// pattern will be eventually used in the future
func (pubsub *PubSub) Subscribe(patten string, c chan *Message) {
	pubsub.subscribers = append(pubsub.subscribers, c)
}
