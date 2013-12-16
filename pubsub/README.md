# PubSub

## Usage

    package main

    import (
      "github.com/dynport/dgtk/pubsub"
      "log"
    )

    func init() {
      log.SetFlags(0)
    }

    type User struct {
      name string
    }

    func main() {
      ps := pubsub.New()
      stringSubscription := ps.Subscribe(func(m string) {
        log.Printf("got string %q", m)
      })
      defer stringSubscription.Close()

      userSubscription := ps.Subscribe(func(u *User) {
        log.Printf("got user %+v", u)
      })
      defer userSubscription.Close()

      ps.Publish("hello")
      ps.Publish("world")

      ps.Publish(&User{name: "Hans"})
      ps.Publish(&User{name: "Meyer"})
    }
