package pubsub

import (
	"fmt"
	"log"
	"reflect"
	"time"
)

type Subscription struct {
	buffer   chan reflect.Value
	finished chan interface{}
	callback reflect.Value
	type_    reflect.Type
	closed   bool
}

const defaultBufferSize = 1000

func (subscription *Subscription) Close() error {
	subscription.closed = true
	close(subscription.buffer)
	timer := time.NewTimer(5 * time.Second)
	select {
	case <-timer.C:
		return fmt.Errorf("timeout waiting for finish")
	case <-subscription.finished:
		return nil
	}
}

func (subscription *Subscription) Matches(v reflect.Value) bool {
	t := v.Type()
	if subscription.type_ == t {
		return true
	}
	if subscription.type_.Kind() == reflect.Interface {
		return t.Implements(subscription.type_)
	}
	return false
}

func (subscription *Subscription) trigger(v reflect.Value) {
	defer func() {
		if r := recover(); r != nil {
			log.Print("PANIC: ", r)
		}
	}()
	subscription.callback.Call([]reflect.Value{v})
}

func (subscription *Subscription) start() {
	subscription.buffer = make(chan reflect.Value, defaultBufferSize)
	subscription.finished = make(chan interface{})
	go func() {
		for value := range subscription.buffer {
			subscription.trigger(value)
		}
		subscription.finished <- nil
	}()
}
