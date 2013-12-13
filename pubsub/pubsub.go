package pubsub

import (
	"fmt"
)

type PubSub struct {
	subscribers []chan *Message
	Stats
}

func (pubsub *PubSub) Publish(key string, payload interface{}) error {
	m := NewMessage(key, payload)
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
