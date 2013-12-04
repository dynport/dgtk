# AMQP Wrapper

This is a simple more golang (use defaults where possible) style wrapper for [github.com/streadway/amqp](http://github.com/streadway/amqp)

## Bind to an exchange (auto-creates the necessary queue)
    package main

    import (
        "log"
        "github.com/dynport/dgtk/amqp"
    )

    func main() {
        connection := &amqp.Connection{
            Address: "amqp://127.0.0.1:5672",
        }
        defer connection.Close()
        queue := &amqp.Queue{
            Name:       "store_metrix",
            AutoDelete: true,
        }
        exchange := &amqp.Exchange{
            Name: "metrix",
        }
        binding := &amqp.Binding{
            Queue:    queue,
            Exchange: exchange,
        }
        e := connection.BindingBind(binding)
        if e != nil {
            return e
        }
        consumer := amqp.Consumer{
            Queue:      queue,
            Connection: connection,
        }
        c, e := consumer.Consume()
        if e != nil {
            log.Fatal(e)
        }
        for del := range c {
          log.Printf("%v", string(del.Body))
        }
    }

