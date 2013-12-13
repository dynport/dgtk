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
			So(s.Publish("hello", "world", nil), ShouldBeNil)
			// sleep for now because the stats update is too slow
			// FIX: use waitFor
			time.Sleep(10 * time.Millisecond)
			So(s.Stats.Received(), ShouldEqual, 1)
			So(s.Stats.Dispatched(), ShouldEqual, 0)
		})

		Convey("Subscribe with consumer", func() {
			s := &PubSub{}
			c := make(chan *Message, 10)
			s.Subscribe("*", c)
			timer := time.NewTicker(1 * time.Second)
			So(s.Publish("hello", "world", nil), ShouldBeNil)
			time.Sleep(10 * time.Millisecond)
			So(s.Stats.Received(), ShouldEqual, 1)
			So(s.Stats.Dispatched(), ShouldEqual, 1)
			var m *Message
			select {
			case <-timer.C:
				t.Fatal("timeout waiting for message")
			case m = <-c:
			}
			So(m.Key, ShouldEqual, "hello")
			So(m.Message, ShouldEqual, "world")
			So(m.CreatedAt.IsZero(), ShouldBeFalse)
			So(m.Level, ShouldEqual, LevelDebug)
		})

		Convey("Subscribe with full message", func() {
			s := &PubSub{}
			c := make(chan *Message, 10)
			s.Subscribe("*", c)
			timer := time.NewTicker(1 * time.Second)
			in := &Message{
				Level:     2,
				CreatedAt: time.Unix(10, 0),
				Duration:  3 * time.Second,
				Payload:   "payload",
			}
			So(s.Publish("hello", "world", in), ShouldBeNil)
			time.Sleep(10 * time.Millisecond)
			So(s.Stats.Received(), ShouldEqual, 1)
			So(s.Stats.Dispatched(), ShouldEqual, 1)
			var m *Message
			select {
			case <-timer.C:
				t.Fatal("timeout waiting for message")
			case m = <-c:
			}
			So(m.Key, ShouldEqual, "hello")
			So(m.Message, ShouldEqual, "world")
			So(m.CreatedAt.Unix(), ShouldEqual, 10)
			So(m.Payload, ShouldEqual, "payload")
			So(m.Level, ShouldEqual, 2)
		})

		Convey("Subscribe without consumer", func() {
			s := &PubSub{}
			c := make(chan *Message, 0)
			s.Subscribe("*", c)
			e := s.Publish("hello", "world", nil)
			So(e, ShouldNotBeNil)
			So(s.SubscribersCount(), ShouldEqual, 1)
		})
	})
}
