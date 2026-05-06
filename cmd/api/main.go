package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/notifylayer/notifylayer/internal/api/server"
	"github.com/notifylayer/notifylayer/internal/db"
	"github.com/notifylayer/notifylayer/internal/queue"
)

func main() {
	ctx := context.Background()

	dsn := mustEnv("DATABASE_URL")
	redisAddr := mustEnv("REDIS_ADDR")
	allowedOrigin := mustEnv("ALLOWED_ORIGIN")
	mustEnv("JWT_SECRET")

	pool, err := db.Connect(ctx, dsn)
	if err != nil {
		slog.Error("database connection failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	q := queue.NewClient(redisAddr)
	defer q.Close() //nolint:errcheck

	addr := envOr("API_ADDR", ":8080")
	srv := &http.Server{
		Addr:         addr,
		Handler:      server.New(pool, q, allowedOrigin),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("api server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "err", err)
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

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
