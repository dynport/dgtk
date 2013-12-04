package amqp

import (
	"github.com/streadway/amqp"
)

type Connection struct {
	Address    string
	connection *amqp.Connection
	channel    *amqp.Channel
}

func (con *Connection) Connection() (c *amqp.Connection, e error) {
	if con.connection != nil {
		return con.connection, nil
	}
	con.connection, e = amqp.Dial(con.Address)
	if e != nil {
		return nil, e
	}
	return con.connection, nil
}

func (con *Connection) Channel() (c *amqp.Channel, e error) {
	if con.channel != nil {
		return con.channel, nil
	}
	connection, e := con.Connection()
	if e != nil {
		return nil, e
	}
	con.channel, e = connection.Channel()
	if e != nil {
		return nil, e
	}
	return con.channel, nil
}

func (con *Connection) Close() {
	if con.channel != nil {
		con.channel.Close()
	}

	if con.connection != nil {
		con.connection.Close()
	}
}

func (con *Connection) BindingBind(binding *Binding) error {
	c, e := con.Channel()
	if e != nil {
		return e
	}
	e = con.QueueDeclare(binding.Queue)
	if e != nil {
		return e
	}
	return c.QueueBind(binding.Queue.Name, binding.Pattern, binding.Exchange.Name, binding.NoWait, binding.Args)
}

func (con *Connection) QueueDeclare(q *Queue) (e error) {
	c, e := con.Channel()
	if e != nil {
		return e
	}

	_, e = c.QueueDeclare(q.Name, q.Durable, q.AutoDelete, q.Exclusive, q.NoWait, q.Args)
	return e
}

type Consumer struct {
	*Connection
	*Queue
	Name      string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
}

func (consumer *Consumer) Consume() (c <-chan amqp.Delivery, e error) {
	ch, e := consumer.Channel()
	if e != nil {
		return nil, e
	}
	return ch.Consume(consumer.Queue.Name, consumer.Name, consumer.AutoAck, consumer.Exclusive, consumer.NoLocal, consumer.NoWait, consumer.Args)
}

type Queue struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

type Exchange struct {
	Name string
}

type Binding struct {
	*Queue
	*Exchange
	Pattern string
	NoWait  bool
	Args    amqp.Table
}
