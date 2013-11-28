package logging

import (
	"bufio"
	"github.com/streadway/amqp"
	"io"
)

type RabbitFeeder struct {
	Address  string // <host>:<port>
	Exchange string
	Ttl      int32
	*amqp.Channel
}

func (feeder *RabbitFeeder) Connection() (con *amqp.Connection, e error) {
	return amqp.Dial("amqp://" + feeder.Address)
}

func (feeder *RabbitFeeder) Feed(in io.Reader) error {
	scanner := bufio.NewScanner(in)
	amqpConnection, e := feeder.Connection()
	if e != nil {
		return e
	}
	defer amqpConnection.Close()
	feeder.Channel, e = amqpConnection.Channel()
	if e != nil {
		return e
	}
	defer feeder.Channel.Close()
	table := amqp.Table{}
	if feeder.Ttl > 0 {
		table["x-message-ttl"] = feeder.Ttl
	}
	e = feeder.Channel.ExchangeDeclare(feeder.Exchange, "fanout", false, false, false, false, table)
	if e != nil {
		return e
	}
	for scanner.Scan() {
		raw := scanner.Text()
		e := feeder.publishLog(raw)
		if e != nil {
			return e
		}
	}
	return nil
}

func (feeder *RabbitFeeder) publishLog(raw string) error {
	line := SyslogLine{}
	if e := line.Parse(raw); e != nil {
		return e
	}
	return feeder.Channel.Publish(feeder.Exchange, line.Host+"."+line.Tag, false, false,
		amqp.Publishing{
			Body: []byte(raw),
		},
	)
}
