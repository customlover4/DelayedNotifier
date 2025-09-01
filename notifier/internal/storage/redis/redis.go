package redis

import (
	"context"
	"delayednotifier/internal/entities/notification"
	"strconv"

	"github.com/go-redis/redis/v8"
	wbfRedis "github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/zlog"
)

type Redis struct {
	rd *wbfRedis.Client
}

func New(addr, password string, db int) *Redis {
	const op = "internal.storage.redis.New"
	r := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	cmd := r.Ping(context.Background())
	if cmd.Err() != nil {
		zlog.Logger.Error().AnErr("err", cmd.Err()).Msg(op)
		panic(cmd.Err())
	}
	return &Redis{
		rd: &wbfRedis.Client{
			Client: r,
		},
	}
}

func (r *Redis) Shutdown() {
	const op = "internal.storage.redis.shutdown"

	err := r.rd.Client.Close()
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
	}
}

func (r *Redis) AddNotification(n notification.Notification) error {
	const op = "internal.storage.redis.AddNotification"

	err := r.rd.Set(context.Background(), strconv.Itoa(int(n.ID)), n)
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
		return err
	}

	return nil
}

func (r *Redis) GetNotification(id int64) (notification.Notification, error) {
	const op = "internal.storage.redis.GetNotification"

	var n notification.Notification
	s, err := r.rd.Get(context.Background(), strconv.Itoa(int(id)))
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
		return n, err
	}
	err = n.UnmarshalBinary([]byte(s))
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
		return n, err
	}
	return n, nil
}

func (r *Redis) DeleteNotification(id int64) (int64, error) {
	const op = "internal.storage.redis.DeleteNotification"

	res, err := r.rd.Del(context.Background(), strconv.Itoa(int(id))).Result()
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
		return 0, err
	}

	return res, nil
}
