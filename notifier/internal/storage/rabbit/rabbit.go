package rabbit

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/zlog"
)

type Rabbit struct {
	ch   *amqp091.Channel
	conn *amqp091.Connection
	q    amqp091.Queue
	ex   string
	key  string
}

func (r *Rabbit) Shoutdown() {
	const op = "internal.storage.rabbit.shutdown"

	err := r.ch.Close()
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
	}
	err = r.conn.Close()
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
	}
}

func New(username, password, addr, queue, ex, key string) *Rabbit {
	r := &Rabbit{}

	conn, err := amqp091.Dial(
		fmt.Sprintf(
			"amqp://%s:%s@%s/", username, password, addr,
		),
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
		queue,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}
	r.q = q

	err = ch.ExchangeDeclare(
		ex,
		"x-delayed-message",
		true,
		false,
		false,
		false,
		amqp091.Table{
			"x-delayed-type": "direct",
		},
	)
	if err != nil {
		panic(err)
	}
	r.ex = ex
	r.key = key

	err = ch.QueueBind(q.Name, key, ex, false, nil)
	if err != nil {
		panic(err)
	}

	return r
}

func (r *Rabbit) Publish(val []byte, d int64) error {
	err := r.ch.Publish(r.ex, r.key, true, false, amqp091.Publishing{
		Headers: amqp091.Table{
			"x-delay": d,
		},
		ContentType: "text/plain",
		Body:        val,
	})
	if err != nil {
		return err
	}

	return nil
}
