package postgres

import (
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

const (
	NotificationTable = "notifications"
)

type Postgres struct {
	db *dbpg.DB
}

func New(host, port, username, password, dbname, sslmode string) *Postgres {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, password, dbname,
	)

	db, err := dbpg.New(connStr, []string{}, &dbpg.Options{
		MaxOpenConns:    20,
		MaxIdleConns:    5,
		ConnMaxLifetime: 2 * time.Hour,
	})
	if err != nil {
		panic(err)
	}
	err = db.Master.Ping()
	if err != nil {
		panic(err)
	}

	return &Postgres{db}
}

func (r *Postgres) Shutdown() {
	const op = "internal.storage.redis.shutdown"

	err := r.db.Master.Close()
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
	}
}
