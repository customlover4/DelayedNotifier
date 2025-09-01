package storage

import (
	"github.com/rabbitmq/amqp091-go"
)

type Messager interface {
	Channel() <-chan amqp091.Delivery
	Shutdown()
}

type Storage struct {
	q Messager
}

func New(q Messager) *Storage {
	return &Storage{
		q: q,
	}
}

func (s *Storage) Receiver() <-chan amqp091.Delivery {
	return s.q.Channel()
}

func (s *Storage) Shutdown() {
	s.q.Shutdown()
}
