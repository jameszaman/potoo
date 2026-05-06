package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/hibiken/asynq"

	"github.com/potoo/potoo/internal/db"
	"github.com/potoo/potoo/internal/queue"
	"github.com/potoo/potoo/internal/worker"
)

func main() {
	ctx := context.Background()

	dsn := mustEnv("DATABASE_URL")
	redisAddr := mustEnv("REDIS_ADDR")

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

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "ERROR: required environment variable %q is not set\n", key)
		os.Exit(1)
	}
	return v
}
