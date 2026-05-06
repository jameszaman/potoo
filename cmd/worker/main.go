package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/hibiken/asynq"

	"github.com/notifylayer/notifylayer/internal/db"
	"github.com/notifylayer/notifylayer/internal/queue"
	"github.com/notifylayer/notifylayer/internal/worker"
)

func main() {
	ctx := context.Background()

	dsn := envOr("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/notify?sslmode=disable")
	redisAddr := envOr("REDIS_ADDR", "localhost:6379")

	pool, err := db.Connect(ctx, dsn)
	if err != nil {
		slog.Error("database connection failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default": 1,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeSendEmail, worker.NewEmailHandler(pool).ProcessTask)

	slog.Info("worker starting", "redis", redisAddr)
	if err := srv.Run(mux); err != nil {
		slog.Error("worker error", "err", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
