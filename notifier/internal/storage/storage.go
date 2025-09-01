package storage

import (
	"database/sql"
	"delayednotifier/internal/entities/notification"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/wb-go/wbf/zlog"
)

var (
	ErrNotFound    = errors.New("not found row")
	ErrNotAffected = errors.New("no one element didn't was affected")
	ErrDontHaveID  = errors.New("can't get id after inserting")
)

type DB interface {
	CreateNotification(n notification.Notification) (int64, error)
	Notification(id int64) (notification.Notification, error)
	UpdateNotificationStatus(status string, id int64) (int64, error)
	DeleteNotification(id int64) (int64, error)
}

type Cache interface {
	AddNotification(n notification.Notification) error
	GetNotification(id int64) (notification.Notification, error)
	DeleteNotification(id int64) (int64, error)
}

type Queue interface {
	Publish(val []byte, d int64) error
}

type Storage struct {
	db DB
	c  Cache
	q  Queue
}

func New(db DB, c Cache, q Queue) *Storage {
	return &Storage{
		db: db,
		c:  c,
		q:  q,
	}
}

func (s *Storage) CreateNotification(n notification.Notification) (int64, error) {
	const op = "internal.storage.CreateNotification"

	id, err := s.db.CreateNotification(n)
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
		return id, err
	}
	if id < 1 {
		return id, ErrDontHaveID
	}
	n.ID = id

	v, err := n.MarshalBinary()
	if err != nil {
		return 0, err
	}

	err = s.q.Publish(v, n.Date.UnixMilli()-time.Now().UnixMilli())
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Storage) UpdateNotificationStatus(status string, id int64) error {
	const op = "internal.storage.UpdateNotification"

	affected, err := s.db.UpdateNotificationStatus(status, id)
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
		return err
	}
	if affected == 0 {
		return ErrNotAffected
	}

	_, err = s.c.DeleteNotification(id)
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	return nil
}

func (s *Storage) GetNotification(id int64) (notification.Notification, error) {
	const op = "internal.storage.GetNotification"

	n, err := s.c.GetNotification(id)
	if err != nil && !errors.Is(err, redis.Nil) {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
		return n, err
	}
	if n.ID != 0 {
		return n, nil
	}

	n, err = s.db.Notification(id)
	if errors.Is(err, sql.ErrNoRows) {
		return n, ErrNotFound
	} else if err != nil {
		return n, err
	}

	err = s.c.AddNotification(n)
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
	}

	return n, nil
}

func (s *Storage) DeleteNotification(id int64) error {
	const op = "internal.storage.DeleteNotification"
	_, err := s.c.DeleteNotification(id)
	if err != nil {
		return err
	}

	affected, err := s.db.DeleteNotification(id)
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
		return err
	}
	if affected == 0 {
		return ErrNotAffected
	}

	return nil
}
