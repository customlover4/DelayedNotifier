package rabbit

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/zlog"
)

type Queue struct {
	conn     *amqp091.Connection
	ch       *amqp091.Channel
	q        amqp091.Queue
	messages <-chan amqp091.Delivery
}

func (r *Queue) Shutdown() {
	const op = "internal.storage.rabbit.Shutdown"
	err := r.ch.Close()
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
	}
	err = r.conn.Close()
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
	}
}

func New(addr, user, password, queue string) *Queue {
	r := &Queue{}
	conn, err := amqp091.Dial(
		fmt.Sprintf("amqp://%s:%s@%s", user, password, addr),
	)
	if err != nil {
		panic(err)
	}
	r.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	r.ch = ch

	q, err := ch.QueueDeclare(
		queue, false, false, false, false, nil,
	)
	if err != nil {
		panic(err)
	}
	r.q = q

	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}
	r.messages = messages

	return r
}

func (r *Queue) Channel() <-chan amqp091.Delivery {
	return r.messages
}
