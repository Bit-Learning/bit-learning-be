.PHONY: run test generate migrate-up migrate-down

ifneq (,$(wildcard .env))
include .env
export APP_ENV HTTP_ADDR DATABASE_URL JWT_SECRET JWT_TTL AUTO_MIGRATE
export OTEL_EXPORTER_OTLP_ENDPOINT OTEL_SERVICE_NAME
endif

run:
	go run ./cmd/api

test:
	go test ./...

generate:
	sqlc generate

migrate-up:
	goose -dir internal/database/migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir internal/database/migrations postgres "$(DATABASE_URL)" down
