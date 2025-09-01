package tests

import (
	"context"
	"database/sql"
	"delayednotifier/internal/entities/notification"
	"delayednotifier/internal/service"
	"delayednotifier/internal/storage"
	"delayednotifier/internal/storage/postgres"
	"delayednotifier/internal/storage/rabbit"
	"delayednotifier/internal/storage/redis"
	"delayednotifier/internal/web/handlers"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/wb-go/wbf/zlog"
)

const (
	DBHost     = "localhost"
	DBUser     = "test"
	DBPassword = "test"
	DBName     = "testdb"

	RabbitUser       = "admin"
	RabbitPassword   = "password"
	RabbitQueue      = "test"
	RabbitExchange   = "test_ex"
	RabbitRoutingKey = "test_rk"

	DBMapped      = "5432"
	RedisMapped   = "6379"
	KafkaMapped   = "9092"
	RabbitMapped1 = "5672"
	RabbitMapped2 = "15672"
)

func SetupTestDB(t *testing.T) testcontainers.Container {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13-alpine",
		ExposedPorts: []string{fmt.Sprintf("%s/tcp", DBMapped)},
		Env: map[string]string{
			"POSTGRES_USER":     DBUser,
			"POSTGRES_PASSWORD": DBPassword,
			"POSTGRES_DB":       DBName,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, DBMapped)
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port.Port(), DBUser, DBPassword, DBName,
	)

	time.Sleep(time.Second * 3)

	oldDB, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	defer oldDB.Close()

	err = oldDB.Ping()
	require.NoError(t, err)

	applyMigrations(t, oldDB)

	return pgContainer
}

func applyMigrations(t *testing.T, db *sql.DB) {
	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "../migrations")

	goose.SetBaseFS(nil)

	if err := goose.Up(db, migrationsDir); err != nil {
		t.Fatalf("Failed to apply migrations: %v", err)
	}

}

func SetupTestRedis(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:7.2-alpine",
		ExposedPorts: []string{fmt.Sprintf("%s/tcp", RedisMapped)},
		Env: map[string]string{
			"MAXMEMORY":        "100MB",
			"MAXMEMORY_POLICY": "volatile-ttl",
		},
		WaitingFor: wait.ForLog("Ready to accept connections"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	time.Sleep(time.Second * 3)

	return redisContainer
}

func SetupTestRabbit(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image: "rabbitmq:latest",
		ExposedPorts: []string{
			fmt.Sprintf("%s/tcp", RabbitMapped1),
			fmt.Sprintf("%s/tcp", RabbitMapped2),
		},
		Env: map[string]string{
			"RABBITMQ_DEFAULT_USER":  RabbitUser,
			"RABBITMQ_DEFAULT_PASS":  RabbitPassword,
			"RABBITMQ_ERLANG_COOKIE": "secret_cookie",
			"RABBITMQ_PLUGINS_DIR":   "/opt/rabbitmq/plugins:/opt/rabbitmq/custom_plugins",
		},
		ConfigModifier: func(c *container.Config) {
			c.User = "rabbitmq"
		},
	}
	rabbitContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	err = rabbitContainer.CopyFileToContainer(ctx,
		"./plugins/rabbitmq_delayed_message_exchange-4.1.0.ez",
		"/opt/rabbitmq/plugins/rabbitmq_delayed_message_exchange-4.1.0.ez",
		0644,
	)
	require.NoError(t, err)

	_, _, err = rabbitContainer.Exec(ctx, []string{
		"rabbitmq-plugins", "enable", "rabbitmq_delayed_message_exchange",
	})
	require.NoError(t, err)

	time.Sleep(time.Second * 4)

	return rabbitContainer
}

func TestNotifyEnd2End(t *testing.T) {
	zlog.Init()

	db := SetupTestDB(t)
	defer func() {
		_ = db.Terminate(context.Background())
	}()
	host, err := db.Host(context.Background())
	require.NoError(t, err)
	port, err := db.MappedPort(context.Background(), DBMapped)
	require.NoError(t, err)

	p := postgres.New(host, port.Port(), DBUser, DBPassword, DBName, "disable")

	cache := SetupTestRedis(t)
	defer func() {
		_ = cache.Terminate(context.Background())
	}()
	host, err = cache.Host(context.Background())
	require.NoError(t, err)
	port, err = cache.MappedPort(context.Background(), RedisMapped)
	require.NoError(t, err)

	rd := redis.New(fmt.Sprintf("%s:%s", host, port.Port()), "", 0)

	queue := SetupTestRabbit(t)
	defer func() {
		_ = queue.Terminate(context.Background())
	}()
	_ = rd
	_ = p
	host, err = queue.Host(context.Background())
	require.NoError(t, err)
	port, err = queue.MappedPort(context.Background(), RabbitMapped1)
	require.NoError(t, err)

	rb := rabbit.New(
		RabbitUser, RabbitPassword, fmt.Sprintf("%s:%s", host, port.Port()),
		RabbitQueue, RabbitExchange, RabbitRoutingKey,
	)

	str := storage.New(p, rd, rb)
	srv := service.New(str)

	gin.SetMode(gin.TestMode)
	g := gin.Default()

	// -------------------- CREATING NOTIFICATION -------------------------
	/*
		send update request and check what notification was added to database
	*/
	g.POST("/notify", handlers.CreateNotify(srv))

	body := `{"message": "hi", "telegram_id": "123", "date": "3000-12-22T15:00:00.000Z"}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost, "/notify", strings.NewReader(body),
	)
	g.ServeHTTP(rr, req)

	t.Log(rr.Body.String())
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	n, err := str.GetNotification(1)
	require.NoError(t, err)
	if n.ID == 0 {
		t.Error("can't find notififcation after adding")
	}
	// --------------------------------------------------------------------

	// -------------------- GETTING NOTIFICATION --------------------------
	/*
		sending get request and check what notification was added to cache
	*/
	g.GET("/notify/:id", handlers.GetNotify(srv))

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(
		http.MethodGet, "/notify/1", nil,
	)
	g.ServeHTTP(rr, req)

	t.Log(rr.Body.String())
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	_, err = rd.GetNotification(n.ID)
	require.NoError(t, err)
	// --------------------------------------------------------------------

	// ------------------ UPDATING NOTIFICATION ---------------------------
	/*
		send request and check what status of notification was updated
	*/
	g.PATCH("/notify/:id", handlers.UpdateNotify(srv))

	body = `{"status": "complete"}`
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(
		http.MethodPatch, "/notify/1", strings.NewReader(body),
	)
	g.ServeHTTP(rr, req)

	t.Log(rr.Body.String())
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	n, err = str.GetNotification(1)
	require.NoError(t, err)
	if n.Status != notification.StatusComplete {
		t.Error("status don't changed after handler")
	}
	// --------------------------------------------------------------------

	// ------------------- DELETING NOTIFICATION --------------------------
	/*
		send request and check what notification was deleted
	*/
	g.DELETE("/notify/:id", handlers.DeleteNotify(srv))

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(
		http.MethodDelete, "/notify/1", nil,
	)
	g.ServeHTTP(rr, req)

	t.Log(rr.Body.String())
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	n, err = str.GetNotification(1)
	if !errors.Is(err, storage.ErrNotFound) {
		t.Error("found notification after deleting")
	}
	// --------------------------------------------------------------------
}
