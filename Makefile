.PHONY: setup dev api web worker db-up db-down migrate generate lint test test-integration fmt clean hooks

-include .env
export

setup:
	go mod download
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	cd apps/web && pnpm install
	pre-commit install

hooks:
	pre-commit install
	pre-commit install --hook-type commit-msg

dev:
	docker compose up -d postgres redis
	make migrate
	make generate

api:
	go run ./cmd/api

worker:
	go run ./cmd/worker

web:
	cd apps/web && pnpm dev

db-up:
	docker compose up -d postgres redis

db-down:
	docker compose down

migrate:
	go run ./cmd/migrate up

generate:
	$(shell go env GOPATH)/bin/oapi-codegen --config api/oapi-codegen.yaml api/openapi.yaml
	$(shell go env GOPATH)/bin/sqlc generate

lint:
	golangci-lint run ./...
	cd apps/web && pnpm lint

fmt:
	go fmt ./...
	cd apps/web && pnpm format

test:
	go test ./...
	cd apps/web && pnpm test

test-integration:
	go test -tags=integration ./...

clean:
	rm -rf tmp bin dist
