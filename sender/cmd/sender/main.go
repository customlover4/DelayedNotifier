package main

import (
	"os"
	"os/signal"
	"sender/internal/service"
	"sender/internal/storage"
	"sender/internal/storage/rabbit"
	"syscall"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.Init()
	cfg := config.New()
	err := cfg.Load(os.Getenv("CONFIG_PATH"))
	if err != nil {
		panic(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	rb := rabbit.New(
		cfg.GetString("rabbit.host"), cfg.GetString("rabbit.username"),
		os.Getenv("RABBIT_PASSWORD"), cfg.GetString("rabbit.queue"),
	)
	str := storage.New(rb)
	srv := service.New(str, cfg.GetString("sender.email_username"))

	go srv.Start()
	zlog.Logger.Info().Msg("start receive messages from queue")

	<-sig
	str.Shutdown()
}
