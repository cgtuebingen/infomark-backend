// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  ComputerGraphics Tuebingen
// Authors: Patrick Wieschollek
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package service

import (
  "fmt"
  "log"

  "github.com/streadway/amqp"
)

type Consumer struct {
  Config *Config

  conn    *amqp.Connection
  channel *amqp.Channel
  done    chan error

  handleFunc func(body []byte) error
}

func NewConsumer(cfg *Config, handleFunc func(body []byte) error) (*Consumer, error) {

  consumer := &Consumer{
    conn:       nil,
    channel:    nil,
    done:       make(chan error),
    handleFunc: handleFunc,

    Config: cfg,
  }

  return consumer, nil
}

func (c *Consumer) Setup() (<-chan amqp.Delivery, error) {

  fmt.Println("-- Connection", c.Config.Connection)
  fmt.Println("-- Exchange", c.Config.Exchange)
  fmt.Println("-- ExchangeType", c.Config.ExchangeType)
  fmt.Println("-- Queue", c.Config.Queue)
  fmt.Println("-- Key", c.Config.Key)
  fmt.Println("-- Tag", c.Config.Tag)

  var err error

  log.Printf("dialing %s", c.Config.Connection)
  c.conn, err = amqp.Dial(c.Config.Connection)
  if err != nil {
    return nil, fmt.Errorf("Dial: %s", err)
  }

  log.Printf("got Connection, getting Channel")
  c.channel, err = c.conn.Channel()
  if err != nil {
    return nil, fmt.Errorf("Channel: %s", err)
  }

  log.Printf("got Channel, declaring Exchange (%s)", c.Config.Exchange)
  if err = c.channel.ExchangeDeclare(
    c.Config.Exchange,     // name of the exchange
    c.Config.ExchangeType, // type
    false,                 // durable
    false,                 // delete when complete
    false,                 // internal
    false,                 // noWait
    nil,                   // arguments
  ); err != nil {
    return nil, fmt.Errorf("Exchange Declare: %s", err)
  }

  log.Printf("declared Exchange, declaring Queue (%s)", c.Config.Queue)
  state, err := c.channel.QueueDeclare(
    c.Config.Queue, // name of the queue
    true,           // durable
    false,          // delete when usused
    false,          // exclusive
    false,          // noWait
    nil,            // arguments
  )
  if err != nil {
    return nil, fmt.Errorf("Queue Declare: %s", err)
  }

  log.Printf("declared Queue (%d messages, %d consumers), binding to Exchange (key '%s')",
    state.Messages, state.Consumers, c.Config.Key)

  if err = c.channel.QueueBind(
    c.Config.Queue,    // name of the queue
    c.Config.Key,      // bindingKey
    c.Config.Exchange, // sourceExchange
    false,             // noWait
    nil,               // arguments
  ); err != nil {
    return nil, fmt.Errorf("Queue Bind: %s", err)
  }

  log.Printf("Queue bound to Exchange, starting Consume (Consumer tag '%s')", c.Config.Tag)
  deliveries, err := c.channel.Consume(
    c.Config.Queue, // name
    c.Config.Tag,   // consumerTag,
    false,          // noAck
    false,          // exclusive
    false,          // noLocal
    false,          // noWait
    nil,            // arguments
  )
  if err != nil {
    return nil, fmt.Errorf("Queue Consume: %s", err)
  }

  return deliveries, nil

}
func (c *Consumer) Shutdown() error {
  // will close() the deliveries channel
  if err := c.channel.Cancel(c.Config.Tag, true); err != nil {
    return fmt.Errorf("Consumer cancel failed: %s", err)
  }

  if err := c.conn.Close(); err != nil {
    return fmt.Errorf("AMQP connection close error: %s", err)
  }

  defer log.Printf("AMQP shutdown OK")

  // wait for handle() to exit
  return <-c.done
}
func (c *Consumer) HandleLoop(deliveries <-chan amqp.Delivery) {
  for d := range deliveries {

    log.Printf(
      "got %dB delivery: [%v] %s",
      len(d.Body),
      d.DeliveryTag,
      d.Body,
    )

    if err := c.handleFunc(d.Body); err != nil {
      d.Ack(false)
    } else {

      d.Ack(true)
    }

  }
  log.Printf("handle: deliveries channel closed")
  c.done <- nil
}