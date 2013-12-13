# PubSub

## Usage
      
    package main

    import (
      "github.com/dynport/dgtk/pubsub"
      "log"
    )

    func main() {
      s := &pubsub.PubSub{}

      // consumers are responsible for providing channels
      // in that case the consumer would buffer 10 messages
      // the pubsub publisher drops messages when consumers are busy
      c := make(chan *pubsub.Message, 10)

      s.Subscribe("*", c) // the pattern is not used at the moment
      s.Publish("hello", nil) // without payload
      s.Publish("hello", "world")
      close(c)
      for m := range c {
        log.Printf("got message %+v", m)
      }
    }
