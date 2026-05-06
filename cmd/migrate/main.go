package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const migrationsDir = "internal/db/migrations"

func main() {
	dsn := envOr("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/notify?sslmode=disable")
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		slog.Error("open db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		slog.Error("set dialect", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := goose.RunContext(ctx, command, db, migrationsDir); err != nil {
		slog.Error("migration failed", "command", command, "err", err)
		os.Exit(1)
	}

	fmt.Printf("migration %q complete\n", command)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
