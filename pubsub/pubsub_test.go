package pubsub

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

type TestUser struct {
	name string
}

func (user *TestUser) String() string {
	return user.name
}

func TestPubSub(t *testing.T) {
	Convey("PubSub", t, func() {
		Convey("Publish", func() {
			s := &PubSub{}
			So(s.Publish("hello"), ShouldBeNil)
			// sleep for now because the stats update is too slow
			// FIX: use waitFor
			time.Sleep(100 * time.Millisecond)
			So(s.Stats.Received(), ShouldEqual, 1)
			So(s.Stats.Dispatched(), ShouldEqual, 0)

			Convey("on closed channel", func() {
				ps := &PubSub{}
				s := ps.Subscribe(func(i interface{}) {
					//
				})
				ps.Publish("something")
				s.Close()
				So(func() { ps.Publish("hello") }, ShouldNotPanic)
			})
		})

		Convey("Subscribe with consumer", func() {
			s := &PubSub{}
			c := make(chan *Message, 10)
			sub := s.Subscribe(func(i *Message) {
				c <- i
			})
			So(s.Publish("hello world"), ShouldBeNil)
			So(s.Publish(&Message{key: "hello world"}), ShouldBeNil)
			time.Sleep(10 * time.Millisecond)
			So(s.Stats.Received(), ShouldEqual, 2)
			So(s.Stats.Dispatched(), ShouldEqual, 1)
			sub.Close()
			var m *Message
			timer := time.NewTicker(1 * time.Second)
			select {
			case <-timer.C:
				t.Fatal("timeout waiting for message")
			case m = <-c:
			}
			So(m.Key(), ShouldEqual, "hello world")
		})

		Convey("Subscribe with interface", func() {
			s := &PubSub{}
			c := make(chan fmt.Stringer, 1)

			sub := s.Subscribe(func(i fmt.Stringer) {
				c <- i
			})
			timer := time.NewTicker(1 * time.Second)
			s.Publish(&TestUser{name: "hans meyer"})
			sub.Close()
			var m fmt.Stringer
			select {
			case <-timer.C:
				t.Fatal("timeout waiting for message")
			case m = <-c:
			}
			So(m.String(), ShouldEqual, "hans meyer")
		})
	})
}

func BenchmarkPublish(b *testing.B) {
	ps := &PubSub{}
	c := make(chan int)
	ready := make(chan int)
	go func() {
		total := 0
		for i := range c {
			total += i
		}
		ready <- total
	}()
	s := ps.Subscribe(func(int) {
		c <- 1
	})
	for i := 0; i < b.N; i++ {
		ps.Publish(10)
		time.Sleep(1 * time.Microsecond)
	}
	s.Close()
	close(c)

	timer := time.NewTicker(5 * time.Second)
	var total int
	select {
	case <-timer.C:
		b.Fatal("timeout waiting for total")
	case total = <-ready:
	}
	b.Logf("%d: %+v", total, ps.Stats)
}

func BenchmarkSubscribe(b *testing.B) {
	s := &PubSub{}
	for i := 0; i < b.N; i++ {
		s.Subscribe(func(string) {})
	}
}
