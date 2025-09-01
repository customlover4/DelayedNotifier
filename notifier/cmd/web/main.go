package main

import (
	"delayednotifier/internal/service"
	"delayednotifier/internal/storage"
	"delayednotifier/internal/storage/postgres"
	"delayednotifier/internal/storage/rabbit"
	"delayednotifier/internal/storage/redis"
	"delayednotifier/internal/web"
	"os/signal"
	"syscall"

	"fmt"
	"os"
	"strconv"

	_ "delayednotifier/docs"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

// @title DelayedNotifier
// @version 0.0.1
// @description Создает отложенные уведомления в очереди
// @host localhost:8080
// @BasePath /

func main() {
	zlog.Init()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	cfg := config.New()
	err := cfg.Load(os.Getenv("CONFIG_PATH"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if os.Getenv("DEBUG") == "false" {
		gin.SetMode(gin.ReleaseMode)
	}

	rb := rabbit.New(
		cfg.GetString("rabbit.username"), os.Getenv("RABBIT_PASSWORD"),
		cfg.GetString("rabbit.host"), cfg.GetString("rabbit.queue"),
		cfg.GetString("rabbit.exchanger"), cfg.GetString("rabbit.routing_key"),
	)
	rdI, err := strconv.Atoi(cfg.GetString("redis.db"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	rd := redis.New(
		cfg.GetString("redis.addr"),
		os.Getenv("REDIS_PASSWORD"),
		rdI,
	)
	db := postgres.New(
		cfg.GetString("postgres.host"), cfg.GetString("postgres.port"),
		cfg.GetString("postgres.username"), os.Getenv("POSTGRES_PASSWORD"),
		cfg.GetString("postgres.dbname"), cfg.GetString("postgres.sslmode"),
	)
	str := storage.New(db, rd, rb)

	srv := service.New(str)

	router := ginext.New()
	router.LoadHTMLGlob("templates/*.html")
	web.SetRoutes(router, srv)

	zlog.Logger.Info().Msg("start listening port")
	go router.Run(
		fmt.Sprintf(":%s", os.Getenv("PORT")),
	)

	<-sig
	zlog.Logger.Info().Msg("shutdown sarting...")
	rb.Shoutdown()
	rd.Shutdown()
	db.Shutdown()
}
