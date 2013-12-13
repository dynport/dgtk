package pubsub

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestPubSub(t *testing.T) {
	Convey("PubSub", t, func() {
		Convey("Publish", func() {
			s := &PubSub{}
			So(s.Publish("hello", nil), ShouldBeNil)
			// sleep for now because the stats update is too slow
			// FIX: use waitFor
			time.Sleep(100 * time.Millisecond)
			So(s.Stats.Received(), ShouldEqual, 1)
			So(s.Stats.Dispatched(), ShouldEqual, 0)
		})

		Convey("Subscribe with consumer", func() {
			s := &PubSub{}
			c := make(chan *Message, 10)
			s.Subscribe("*", c)
			timer := time.NewTicker(1 * time.Second)
			So(s.Publish("hello", nil), ShouldBeNil)
			time.Sleep(10 * time.Millisecond)
			So(s.Stats.Received(), ShouldEqual, 1)
			So(s.Stats.Dispatched(), ShouldEqual, 1)
			var m *Message
			select {
			case <-timer.C:
				t.Fatal("timeout waiting for message")
			case m = <-c:
			}
			So(m.Key(), ShouldEqual, "hello")
		})

		Convey("Subscribe with full message", func() {
			s := &PubSub{}
			c := make(chan *Message, 10)
			s.Subscribe("*", c)
			timer := time.NewTicker(1 * time.Second)
			So(s.Publish("hello", "payload"), ShouldBeNil)
			time.Sleep(10 * time.Millisecond)
			So(s.Stats.Received(), ShouldEqual, 1)
			So(s.Stats.Dispatched(), ShouldEqual, 1)
			var m *Message
			select {
			case <-timer.C:
				t.Fatal("timeout waiting for message")
			case m = <-c:
			}
			So(m.Key(), ShouldEqual, "hello")
			So(m.CreatedAt(), ShouldNotBeNil)
			So(m.Payload(), ShouldEqual, "payload")
		})

		Convey("Subscribe without consumer", func() {
			s := &PubSub{}
			c := make(chan *Message, 0)
			s.Subscribe("*", c)
			e := s.Publish("hello", nil)
			So(e, ShouldNotBeNil)
			So(s.SubscribersCount(), ShouldEqual, 1)
		})
	})
}
