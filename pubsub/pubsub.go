package pubsub

import (
	"fmt"
	"reflect"
)

func New() *PubSub {
	return &PubSub{}
}

type PubSub struct {
	subscriptions []*Subscription
	Stats
}

func (pubsub *PubSub) Publish(i interface{}) error {
	pubsub.Stats.MessageReceived()
	var e error
	value := reflect.ValueOf(i)
	for _, s := range pubsub.subscriptions {
		if !s.closed && s.Matches(value) {
			select {
			case s.buffer <- value:
				pubsub.Stats.MessageDispatched()
			default:
				e = fmt.Errorf("unable to publish to %v", s)
			}
		}
	}
	return e
}

func (pubsub *PubSub) Subscribe(i interface{}) *Subscription {
	value := reflect.ValueOf(i)
	type_ := reflect.TypeOf(i)
	if type_.Kind() != reflect.Func || type_.NumIn() != 1 {
		panic("you must provide a callback with exactly one argument like func(m *Message) {}")
	}

	s := &Subscription{
		callback: value,
		type_:    type_.In(0),
	}
	pubsub.subscriptions = append(pubsub.subscriptions, s)
	s.start()
	return s
}

func (pubsub *PubSub) SubscribersCount() int {
	return len(pubsub.subscriptions)
}

//  could not be bother to fight with locking just now
// func (pubsub *PubSub) Unsubscribe(s *Subscription) {
// implement me
// }
